package manifests

import (
	appsv1 "k8s.io/api/apps/v1"

	"github.com/drone/envsubst"

	"github.com/fromanirh/deployer/pkg/images"
)

func UpdateSchedulerPluginDeployment(dp *appsv1.Deployment) *appsv1.Deployment {
	ret := dp.DeepCopy()
	ret.Spec.Template.Spec.Containers[0].Image = images.SchedulerPluginImage
	return ret
}

func UpdateResourceTopologyExporterDaemonSet(ds *appsv1.DaemonSet) *appsv1.DaemonSet {
	ret := ds.DeepCopy()
	// TODO: better match by name than assume container#0 is RTE proper (not minion)
	ret.Spec.Template.Spec.Containers[0].Image = images.ResourceTopologyExporterImage
	ret.Spec.Template.Spec.Containers[0].Command = UpdateResourceTopologyExporterCommand(ds.Spec.Template.Spec.Containers[0].Command)
	return ret
}

func UpdateResourceTopologyExporterCommand(args []string) []string {
	vars := map[string]string{
		"RTE_POLL_INTERVAL": "10s",
	}
	res := []string{}
	for _, arg := range args {
		newArg, err := envsubst.Eval(arg, func(key string) string { return vars[key] })
		if err != nil {
			// TODO log?
			continue
		}
		res = append(res, newArg)
	}
	return res
}
