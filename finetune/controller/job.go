package controller

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket" // swagger embed files
	"github.com/opensourceways/foundation-model-server/allerror"
	commonctl "github.com/opensourceways/foundation-model-server/common/controller"
	"github.com/opensourceways/foundation-model-server/common/controller/middleware"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/pointer"
)

const headerSecret = "FINETUNE-SECRET"

var (
	kubeconfig string
	namespace  string
	tokens     []string
	image      string
	clientset  *kubernetes.Clientset
)

type JobInfo struct {
	JobName   string            `json:"jobName,omitempty"`
	Username  string            `json:"username"`
	Dataset   string            `json:"dataset"`
	Model     string            `json:"model"`
	CreatedAt string            `json:"created_at,omitempty"`
	Status    string            `json:"status,omitempty"`
	Parameter map[string]string `json:"parameter"`
}

func readLinesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	return lines, nil
}

func Init(n, k, i, t string) error {
	kubeconfig = k
	namespace = n
	image = i

	// 创建 Kubernetes 客户端
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}
	defer os.Remove(kubeconfig)

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	if tokens, err = readLinesFromFile(t); err != nil {
		return err
	}
	defer os.Remove(t)

	return nil
}

// 注册路由
func RegisterRoutes(router *gin.RouterGroup) {
	m := middleware.AccessTokenChecking()
	// 创建作业
	router.POST("/v1/job", m, createJob)
	// 删除作业
	router.DELETE("/v1/job/:jobname", m, deleteJob)

	router.GET("/v1/log/:jobname", m, getJobLogs)
	// 获取所有作业
	router.GET("/v1/job", m, listJobs)
}

//	@Title			List
//	@Description	list jobs
//	@Tags			Finetune
//	@Success		200	{object}		[]JobInfo
//	@Failure		500	system_error	system	error
//	@Router			/v1/job [get]
func listJobs(c *gin.Context) {
	// 使用 clientset 进行操作
	jobList, err := clientset.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		commonctl.SendFailedResp(c, err)
		logrus.Error(err.Error())
		return
	}

	// 构造作业信息列表
	jobInfos := make([]JobInfo, 0)
	for _, job := range jobList.Items {
		var status string
		if len(job.Status.Conditions) == 0 && job.Status.Active > 0 {
			// 当条件列表为空且有活动的副本时，将作业状态设置为"Running"
			status = "Running"
		} else if len(job.Status.Conditions) == 0 {
			// 当条件列表为空且没有活动的副本时，将作业状态设置为"Pending"
			status = "Pending"
		} else {
			// 查找具有最新时间戳的条件
			latestCondition := job.Status.Conditions[len(job.Status.Conditions)-1]
			status = string(latestCondition.Type)
		}

		params, err := getEnvs(&job, true)
		if err != nil {
			commonctl.SendFailedResp(c, err)
			logrus.Error(err.Error())
			return
		}

		jobInfo := JobInfo{
			JobName:   job.Name,
			Username:  job.Labels["create_by"],
			Dataset:   job.Labels["data"],
			Model:     job.Labels["model"],
			CreatedAt: job.CreationTimestamp.Format(time.RFC3339),
			Status:    status,
			Parameter: params,
		}
		jobInfos = append(jobInfos, jobInfo)
	}

	// 返回作业信息列表
	c.JSON(http.StatusOK, jobInfos)
}

// 查询作业状态
func getJobStatus(c *gin.Context) {
	// 从路径参数中获取作业名称和命名空间
	jobName := c.Param("jobname")

	// 使用 clientset 进行操作
	job, err := clientset.BatchV1().Jobs(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		logrus.Error(err.Error())
		return
	}

	// 返回作业状态
	c.JSON(http.StatusOK, JobInfo{
		JobName:   job.Name,
		Username:  job.GetObjectMeta().GetLabels()["create_by"],
		Dataset:   job.GetObjectMeta().GetLabels()["data"],
		Model:     job.GetObjectMeta().GetLabels()["model"],
		Parameter: make(map[string]string),
		CreatedAt: job.CreationTimestamp.Format(time.RFC3339),
		Status:    "Running",
	})
}

func getEnvs(job *batchv1.Job, skipSecret bool) (envs map[string]string, err error) {
	envs = make(map[string]string)
	// 提取 Pod 的环境变量
	for _, container := range job.Spec.Template.Spec.Containers {
		for _, envVar := range container.Env {
			// 不返回secret
			if envVar.Name == "SECRET" && skipSecret {
				continue
			}
			envs[strings.ToLower(envVar.Name)] = envVar.Value
		}
	}
	return
}

//	@Summary		Create
//	@Description	create finetune
//	@Tags			Finetune
//	@Param			body	body	JobInfo	true	"body of creating finetune"
//	@Accept			json
//	@Success		201	{object}		JobInfo
//	@Failure		500	system_error	system	error
//	@Router			/v1/job [post]
func createJob(c *gin.Context) {
	// 从请求中获取作业相关参数
	var jobInfo JobInfo

	// 解析JSON请求体
	if err := c.ShouldBindJSON(&jobInfo); err != nil {
		commonctl.SendBadRequestBody(c, err)
		logrus.Error(err.Error())
		return
	}
	logrus.Infof("username: %s dataset: %s model: %s parameter: %v", jobInfo.Username, jobInfo.Dataset, jobInfo.Model, jobInfo.Parameter)

	jobInfo.Parameter["secret"] = c.GetHeader(headerSecret)
	jobInfo.Parameter["model_name"] = jobInfo.Model
	jobInfo.Parameter["dataset"] = jobInfo.Dataset
	jobInfo.Parameter["npu_number"] = "4"

	// 创建作业对象
	job, err := doCreateJob(clientset, jobInfo.Username, jobInfo.Dataset, jobInfo.Model, &jobInfo.Parameter, 120)
	if err != nil {
		commonctl.SendFailedResp(c, err)
		logrus.Error(err.Error())
		return
	}

	params, err := getEnvs(job, true)
	if err != nil {
		commonctl.SendFailedResp(c, err)
		logrus.Error(err.Error())
		return
	}

	// 返回创建成功的响应
	c.JSON(http.StatusOK, JobInfo{
		JobName:   job.Name,
		Username:  job.GetObjectMeta().GetLabels()["create_by"],
		Dataset:   job.GetObjectMeta().GetLabels()["data"],
		Model:     job.GetObjectMeta().GetLabels()["model"],
		Parameter: params,
		CreatedAt: job.CreationTimestamp.Format(time.RFC3339),
		Status:    "Running",
	})
}

//	@Summary		get a websocket to watch a finetune log
//	@Description	watch single finetune
//	@Tags			Finetune
//	@Param			jobname	path	string	true	"finetune id"
//	@Accept			json
//	@Success		200	{object}		string
//	@Failure		500	system_error	system	error
//	@Router			/v1/log/{jobname} [get]
func getJobLogs(c *gin.Context) {
	// 从路径参数中获取作业名称和命名空间
	jobName := c.Param("jobname")
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许跨域访问
		},
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		commonctl.SendFailedResp(c, err)
		logrus.Error(err.Error())
		return
	}

	defer ws.Close()
	if err := doWatchJob(ws, clientset, namespace, jobName); err != nil {
		commonctl.SendFailedResp(c, err)
		logrus.Error(err.Error())
	}
}

func doWatchJob(ws *websocket.Conn, clientset *kubernetes.Clientset, namespacm, jobName string) error {

	ctx := context.TODO()
	// 获取pod以便获取日志
	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "job-name=" + jobName,
	})
	if err != nil {
		return err
	}

	for _, pod := range podList.Items {
		req := clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
			Follow: true, // 实时跟踪日志
		})
		podLogs, err := req.Stream(ctx)
		if err != nil {
			return err
		}
		defer podLogs.Close()

		// 读取日志内容
		buf := make([]byte, 4096)
		for {
			n, err := podLogs.Read(buf)
			if err != nil {
				break
			}
			if n > 0 {
				if err := ws.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
					break
				}
			}
		}
	}

	return nil
}

//	@Summary		Delete
//	@Description	delete finetune
//	@Tags			Finetune
//	@Param			jobname	path	string	true	"finetune id"
//	@Accept			json
//	@Success		200
//	@Failure		500	system_error	system	error
//	@Router			/v1/job/{jobname} [delete]
func deleteJob(c *gin.Context) {
	// 从路径参数中获取作业名称
	jobname := c.Param("jobname")
	secret := c.GetHeader(headerSecret)

	err := checkDeletePerm(clientset, jobname, namespace, secret)
	if err != nil {
		commonctl.SendFailedResp(c, err)
		logrus.Error(err.Error())
		return
	}

	err = doDeleteJob(clientset, jobname, namespace)
	if err != nil {
		commonctl.SendFailedResp(c, err)
		logrus.Error(err.Error())
		return
	}

	// 返回删除成功的响应
	c.JSON(http.StatusOK, gin.H{
		"msg": fmt.Sprintf("Job %s deleted", jobname),
	})
}

// 创建环境变量列表
func createEnvVars(envVars *map[string]string) []corev1.EnvVar {
	var env []corev1.EnvVar
	for key, value := range *envVars {
		env = append(env, corev1.EnvVar{
			Name:  strings.ToUpper(key),
			Value: value,
		})
	}
	return env
}

func doCreateJob(clientset *kubernetes.Clientset, username, dataset, model string, parameter *map[string]string, timeout int) (jobObj *batchv1.Job, err error) {
	jobName := uuid.New().String()

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels: map[string]string{
				"create_by": username,
				"model":     model,
				"data":      dataset,
				"parameter": "",
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    jobName,
							Image:   image,
							Command: []string{"/bin/bash", "-i", "/root/run_finetune.sh"},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "model",
									MountPath: "/opt/" + model,
								},
								corev1.VolumeMount{
									Name:      "dataset",
									MountPath: "/opt/" + dataset + ".json",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"huawei.com/Ascend910": resourceQuantity("4"),
								},
								Limits: corev1.ResourceList{
									"huawei.com/Ascend910": resourceQuantity("4"),
								},
							},
							Env: createEnvVars(parameter),
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "model",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/data/disk1/model/" + model,
									Type: (*corev1.HostPathType)(pointer.String(string(corev1.HostPathDirectory))),
								},
							},
						},
						{
							Name: "dataset",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/data/disk1/dataset/" + dataset + ".json",
									Type: (*corev1.HostPathType)(pointer.String(string(corev1.HostPathFile))),
								},
							},
						},
					},
				},
			},
			BackoffLimit: pointer.Int32(4),
		},
	}

	jobObj, err = clientset.BatchV1().Jobs(namespace).Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		return
	}

	timeoutTime := time.Now().Add(time.Duration(timeout) * time.Second)

	var pods *corev1.PodList
	// 等待 Pod 创建成功
	for {
		pods, err = clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=%s", jobName),
		})
		if err != nil {
			return
		}

		if len(pods.Items) > 0 {
			// 检查所有 Pod 的状态
			allRunning := true
			for _, pod := range pods.Items {
				if pod.Status.Phase != corev1.PodRunning {
					allRunning = false
					break
				}
			}

			if allRunning {
				// 所有 Pod 都处于 Running 状态
				return
			}
		}

		if time.Now().After(timeoutTime) {
			// 删除 Job 和 Pod
			deletePolicy := metav1.DeletePropagationForeground
			clientset.BatchV1().Jobs(namespace).Delete(context.TODO(), jobName, metav1.DeleteOptions{
				PropagationPolicy: &deletePolicy,
			})

			for _, pod := range pods.Items {
				clientset.CoreV1().Pods(namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{
					PropagationPolicy: &deletePolicy,
				})
			}

			return nil, allerror.New(allerror.ErrorCodeReqTimeout, "timeout waiting for job running")
		}
		// 等待一段时间后重新检查
		time.Sleep(5 * time.Second)
	}
}

func checkDeletePerm(clientset *kubernetes.Clientset, jobName, namespace, secret string) error {
	if secret == "" {
		return allerror.New(allerror.ErrorPermissionDeny, "Permission deny, empty secret")
	}
	// 删除job
	job, err := clientset.BatchV1().Jobs(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	params, err := getEnvs(job, false)
	if err != nil {
		return err
	}
	// check secret
	if s, ok := params["secret"]; ok && s != secret {
		return allerror.New(allerror.ErrorPermissionDeny, "Permission deny, you can't delete the job created by other")
	}
	return nil
}

func doDeleteJob(clientset *kubernetes.Clientset, jobName, namespace string) error {
	// 删除job
	err := clientset.BatchV1().Jobs(namespace).Delete(context.TODO(), jobName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	// 删除相关的 Pod
	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "job-name=" + jobName,
	})
	if err != nil {
		return err
	}

	var pods []string
	for _, pod := range podList.Items {
		err = clientset.CoreV1().Pods(namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		pods = append(pods, pod.Name)
	}
	// 等待job删除成功
	for {
		_, err := clientset.BatchV1().Jobs(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
		if err != nil {
			if isNotFound(err) {
				break
			} else {
				return err
			}
		}

		time.Sleep(1 * time.Second)
	}
	// 等待pod删除成功
	for _, pod := range pods {
		podDeleted := false
		for !podDeleted {
			_, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), pod, metav1.GetOptions{})
			if err != nil {
				if isNotFound(err) {
					podDeleted = true
				} else {
					break
				}
			}

			time.Sleep(1 * time.Second)
		}
	}
	return nil
}

// helpers
func isNotFound(err error) bool {
	if statusError, ok := err.(*errors.StatusError); ok {
		if statusError.ErrStatus.Reason == metav1.StatusReasonNotFound {
			return true
		}
	}
	return false
}

func resourceQuantity(quantity string) resource.Quantity {
	q, err := resource.ParseQuantity(quantity)
	if err != nil {
		panic(err)
	}
	return q
}
