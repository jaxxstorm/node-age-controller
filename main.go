/*
.
*/

package main

import (
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/jaxxstorm/node-age-controller/controllers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	dryRun               = kingpin.Flag("dry-run", "Don't operate on nodes, only log what would happen").Envar("DRY_RUN").Bool()
	development          = kingpin.Flag("development", "Enable development logging").Bool()
	enableLeaderElection = kingpin.Flag("enable-leader-election", "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.").Bool()
	metricsAddr          = kingpin.Flag("metrics-addr", "The address the metric endpoint binds to").Default(":8080").String()
	maxNodes             = kingpin.Flag("max-nodes", "The max number of nodes that can be cordoned at one time").Envar("MAX_NODES").Default("3").Int()
	maxNodeAge           = kingpin.Flag("max-node-age", "How old is allowed to be before we attempt to cordon it").Envar("MAX_NODE_AGE").Default("720h").Duration()
	minAvailableNodes    = kingpin.Flag("min-available-nodes", "How many nodes must be uncordoned before we attempt to cordon a node").Envar("MIN_AVAILABLE_NODES").Default("3").Int()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = corev1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	kingpin.Parse()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		if *development {
			o.Development = true
		}
	}))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: *metricsAddr,
		LeaderElection:     *enableLeaderElection,
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.NodeReconciler{
		Client:            mgr.GetClient(),
		Log:               ctrl.Log.WithName("controllers").WithName("Node"),
		Recorder:          mgr.GetEventRecorderFor("node-controller"),
		Scheme:            mgr.GetScheme(),
		DryRun:            *dryRun,
		MaxNodes:          *maxNodes,
		MaxNodeAge:        *maxNodeAge,
		MinAvailableNodes: *minAvailableNodes,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Node")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
