/*
.
*/

package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"

	// This is to allow oidc auth to work when testing locally
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

const ignoreAnnotation = "age.briggs.io/ignore"

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	DryRun   bool
}

// Reconcile reconciles a node
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
func (r *NodeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("node", req.NamespacedName)

	node := &corev1.Node{}

	if err := r.Client.Get(ctx, req.NamespacedName, node); err != nil {
		log.Error(err, "unable to fetch node")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, err
	}

	// if it's a master node, do nothing
	if isMaster(node) {
		log.Info("Ignoring master node")
		return ctrl.Result{}, nil
	}

	if nodeAnnotated(node) {
		log.Info("Node is annotated, ignoring")
		return ctrl.Result{}, nil
	}

	age := time.Since(node.ObjectMeta.CreationTimestamp.Time)
	name := node.ObjectMeta.Name
	log.WithValues("Name", name, "Age", age).V(1).Info("Checking node age")

	// get the desired node age
	desiredAge, err := time.ParseDuration("2160h")

	if err != nil {
		log.Error(err, "Error parsing desired age")
		return ctrl.Result{}, err
	}

	if age > desiredAge {
		if nodeCordoned(node) {
			log.WithValues("Name", name, "Age", age).Info("Node is already cordoned")
		} else {
			log.WithValues("Name", name, "Age", age).Info("Cordoning node")
			if r.DryRun {
				log.WithValues("Name", name, "Age", age).Info("DryRun Enabled, not cordoning node")
				return ctrl.Result{}, nil
			}
			updatedNode := cordonNode(node)
			err := r.Client.Update(context.TODO(), updatedNode)

			if err != nil {
				return ctrl.Result{}, err
			}
			//r.Recorder.Eventf(updatedNode, corev1.EventTypeNormal, "TaintsChanged", "Added Taint briggs.io/aged")
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager Adds the controller to the manager
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}

func cordonNode(node *corev1.Node) *corev1.Node {
	// takes the node option, and returns a copy of it with the
	// "Unschedulable" set to true
	copy := node.DeepCopy()
	copy.Spec.Unschedulable = true
	return copy
}

// loop through the taints and check for the presence of a master taint key
func isMaster(node *corev1.Node) bool {

	for _, taint := range node.Spec.Taints {
		if taint.Key == "node-role.kubernetes.io/master" {
			return true
		}
	}
	return false
}

// check the Unscheduleable field to determine if node is already cordoned
func nodeCordoned(node *corev1.Node) bool {

	if node.Spec.Unschedulable {
		return true
	}
	return false
}

// check the Annotations to see if this should be ignored
func nodeAnnotated(node *corev1.Node) bool {

	for k, v := range node.ObjectMeta.Annotations {
		if k == ignoreAnnotation {
			if v == "true" {
				return true
			} else {
				return false
			}
		}
	}
	return false

}
