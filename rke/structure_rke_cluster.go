package rke

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/rancher/rke/cluster"
	rancher "github.com/rancher/types/apis/management.cattle.io/v3"
	apiserverconfigv1 "k8s.io/apiserver/pkg/apis/config/v1"
)

// Flatteners

func flattenRKEClusterFlag(d *schema.ResourceData, in *cluster.ExternalFlags) {
	if in == nil {
		return
	}

	d.Set("update_only", in.UpdateOnly)
	d.Set("disable_port_check", in.DisablePortCheck)
	d.Set("dind", in.DinD)
	d.Set("custom_certs", in.CustomCerts)
	if len(in.CertificateDir) > 0 {
		d.Set("cert_dir", in.CertificateDir)
	}
}

func flattenRKECluster(d *schema.ResourceData, in *cluster.Cluster) error {
	if in == nil {
		return nil
	}

	var err error
	if in.AddonJobTimeout > 0 {
		d.Set("addon_job_timeout", int(in.AddonJobTimeout))
	}

	if v, ok := d.Get("addons").(string); len(in.Addons) > 0 && ok && len(v) > 0 {
		d.Set("addons", in.Addons)
	}

	if v, ok := d.Get("addons_include").(string); len(in.AddonsInclude) > 0 && ok && len(v) > 0 {
		d.Set("addons_include", in.AddonsInclude)
	}

	err = d.Set("authentication", flattenRKEClusterAuthentication(in.Authentication))
	if err != nil {
		return err
	}

	err = d.Set("authorization", flattenRKEClusterAuthorization(in.Authorization))
	if err != nil {
		return err
	}

	err = d.Set("bastion_host", flattenRKEClusterBastionHost(in.BastionHost))
	if err != nil {
		return err
	}

	if v, ok := d.Get("cloud_provider").([]interface{}); ok && len(v) > 0 {
		err = d.Set("cloud_provider", flattenRKEClusterCloudProvider(in.CloudProvider, v))
		if err != nil {
			return err
		}
	}

	if len(in.ClusterName) > 0 {
		d.Set("cluster_name", in.ClusterName)
	}

	if in.DNS != nil {
		err := d.Set("dns", flattenRKEClusterDNS(in.DNS))
		if err != nil {
			return err
		}
	}

	d.Set("dind", in.DinD)

	d.Set("ignore_docker_version", *in.IgnoreDockerVersion)

	err = d.Set("ingress", flattenRKEClusterIngress(in.Ingress))
	if err != nil {
		return err
	}

	if len(in.Version) > 0 {
		d.Set("kubernetes_version", in.Version)
	}

	err = d.Set("monitoring", flattenRKEClusterMonitoring(in.Monitoring))
	if err != nil {
		return err
	}

	err = d.Set("network", flattenRKEClusterNetwork(in.Network))
	if err != nil {
		return err
	}

	if v, ok := d.Get("nodes").([]interface{}); in.Nodes != nil && !in.DinD && ok && len(v) > 0 {
		nodes := flattenRKEClusterNodes(in.Nodes, v)
		err := d.Set("nodes", nodes)
		if err != nil {
			return err
		}
	}

	if len(in.PrefixPath) > 0 {
		d.Set("prefix_path", in.PrefixPath)
	}

	if v, ok := d.Get("private_registries").([]interface{}); in.PrivateRegistries != nil && ok && len(v) > 0 {
		err := d.Set("private_registries", flattenRKEClusterPrivateRegistries(in.PrivateRegistries))
		if err != nil {
			return err
		}
	}

	err = d.Set("restore", flattenRKEClusterRestore(in.Restore))
	if err != nil {
		return err
	}

	if v, ok := d.Get("rotate_certificates").([]interface{}); in.RotateCertificates != nil && ok && len(v) > 0 {
		err := d.Set("rotate_certificates", flattenRKEClusterRotateCertificates(in.RotateCertificates))
		if err != nil {
			return err
		}
	}

	if v, ok := d.Get("services").([]interface{}); ok {
		services, err := flattenRKEClusterServices(in.Services, v)
		if err != nil {
			return err
		}
		err = d.Set("services", services)
		if err != nil {
			return err
		}
	}

	d.Set("ssh_agent_auth", in.SSHAgentAuth)

	if len(in.SSHCertPath) > 0 {
		d.Set("ssh_cert_path", in.SSHCertPath)
	}

	if len(in.SSHKeyPath) > 0 {
		d.Set("ssh_key_path", in.SSHKeyPath)
	}

	// computed values
	d.Set("api_server_url", "") // nolint
	if in.ControlPlaneHosts != nil && len(in.ControlPlaneHosts) > 0 {
		apiServerURL := fmt.Sprintf("https://" + in.ControlPlaneHosts[0].Address + ":6443")
		d.Set("api_server_url", apiServerURL)
	}

	caCrt, clientCrt, clientKey, certificates := flattenRKEClusterCertificates(in.Certificates)
	d.Set("ca_crt", caCrt)          // nolint
	d.Set("client_cert", clientCrt) // nolint
	d.Set("client_key", clientKey)  // nolint
	d.Set("certificates", certificates)
	d.Set("kube_admin_user", rkeClusterCertificatesKubeAdminCertName)
	d.Set("cluster_domain", in.ClusterDomain)        // nolint
	d.Set("cluster_cidr", in.ClusterCIDR)            // nolint
	d.Set("cluster_dns_server", in.ClusterDNSServer) // nolint

	err = d.Set("etcd_hosts", flattenRKEClusterNodesComputed(in.EtcdHosts))
	if err != nil {
		return err
	}

	err = d.Set("control_plane_hosts", flattenRKEClusterNodesComputed(in.ControlPlaneHosts))
	if err != nil {
		return err
	}

	err = d.Set("worker_hosts", flattenRKEClusterNodesComputed(in.WorkerHosts))
	if err != nil {
		return err
	}

	err = d.Set("inactive_hosts", flattenRKEClusterNodesComputed(in.InactiveHosts))
	if err != nil {
		return err
	}

	err = d.Set("running_system_images", flattenRKEClusterSystemImages(in.SystemImages))
	if err != nil {
		return err
	}

	err = d.Set("upgrade_strategy", flattenRKEClusterNodeUpgradeStrategy(in.UpgradeStrategy))
	if err != nil {
		return err
	}

	return nil
}

// Expanders

func expandRKECluster(in *schema.ResourceData) (string, *rancher.RancherKubernetesEngineConfig, error) {
	if in == nil {
		return "", nil, nil
	}

	obj := &rancher.RancherKubernetesEngineConfig{}

	if v, ok := in.Get("cluster_yaml").(string); ok && len(v) > 0 {
		var err error
		obj, err = cluster.ParseConfig(v)
		if err != nil {
			return "", nil, err
		}
	}

	if v, ok := in.Get("addon_job_timeout").(int); ok && v > 0 {
		obj.AddonJobTimeout = v
	}

	if v, ok := in.Get("addons").(string); ok && len(v) > 0 {
		obj.Addons = v
	}

	if v, ok := in.Get("addons_include").([]interface{}); ok && len(v) > 0 {
		obj.AddonsInclude = toArrayString(v)
	}

	if v, ok := in.Get("authentication").([]interface{}); ok && len(v) > 0 {
		obj.Authentication = expandRKEClusterAuthentication(v)
	}

	if v, ok := in.Get("authorization").([]interface{}); ok && len(v) > 0 {
		obj.Authorization = expandRKEClusterAuthorization(v)
	}

	if v, ok := in.Get("bastion_host").([]interface{}); ok && len(v) > 0 {
		obj.BastionHost = expandRKEClusterBastionHost(v)
	}

	if v, ok := in.Get("cloud_provider").([]interface{}); ok && len(v) > 0 {
		obj.CloudProvider = expandRKEClusterCloudProvider(v)
	}

	if v, ok := in.Get("cluster_name").(string); ok && len(v) > 0 {
		obj.ClusterName = v
	}

	if v, ok := in.Get("dns").([]interface{}); ok && len(v) > 0 {
		obj.DNS = expandRKEClusterDNS(v)
	}

	if v, ok := in.Get("ignore_docker_version").(bool); ok {
		obj.IgnoreDockerVersion = &v
	}

	if v, ok := in.Get("ingress").([]interface{}); ok && len(v) > 0 {
		obj.Ingress = expandRKEClusterIngress(v)
	}

	if v, ok := in.Get("kubernetes_version").(string); ok && len(v) > 0 {
		obj.Version = v
	}

	if v, ok := in.Get("monitoring").([]interface{}); ok && len(v) > 0 {
		obj.Monitoring = expandRKEClusterMonitoring(v)
	}

	if v, ok := in.Get("network").([]interface{}); ok && len(v) > 0 {
		obj.Network = expandRKEClusterNetwork(v)
	}

	if v, ok := in.Get("nodes").([]interface{}); ok && len(v) > 0 {
		obj.Nodes = expandRKEClusterNodes(v)
	}

	if v, ok := in.Get("prefix_path").(string); ok && len(v) > 0 {
		obj.PrefixPath = v
	}

	if v, ok := in.Get("private_registries").([]interface{}); ok && len(v) > 0 {
		obj.PrivateRegistries = expandRKEClusterPrivateRegistries(v)
	}

	if v, ok := in.Get("restore").([]interface{}); ok && len(v) > 0 {
		obj.Restore = expandRKEClusterRestore(v)
	}

	if v, ok := in.Get("rotate_certificates").([]interface{}); ok && len(v) > 0 {
		obj.RotateCertificates = expandRKEClusterRotateCertificates(v)
	}

	if v, ok := in.Get("ssh_agent_auth").(bool); ok && v {
		obj.SSHAgentAuth = v
	}

	if v, ok := in.Get("ssh_cert_path").(string); ok && len(v) > 0 {
		obj.SSHCertPath = v
	}

	if v, ok := in.Get("ssh_key_path").(string); ok && len(v) > 0 {
		obj.SSHKeyPath = v
	}

	if v, ok := in.Get("system_images").([]interface{}); ok && len(v) > 0 {
		obj.SystemImages = expandRKEClusterSystemImages(v)
	}

	if v, ok := in.Get("upgrade_strategy").([]interface{}); ok {
		obj.UpgradeStrategy = expandRKEClusterNodeUpgradeStrategy(v)
	}

	if v, ok := in.Get("services").([]interface{}); ok && len(v) > 0 {
		services, err := expandRKEClusterServices(v)
		if err != nil {
			return "", nil, err
		}
		obj.Services = services
	}

	if v, ok := in.Get("dind").(bool); ok && v {
		if obj.Services.Kubeproxy.ExtraArgs == nil {
			obj.Services.Kubeproxy.ExtraArgs = make(map[string]string)
		}
		obj.Services.Kubeproxy.ExtraArgs["conntrack-max-per-core"] = "0"
	}

	objYml, err := patchRKEClusterYaml(in, obj)
	if err != nil {
		return "", nil, fmt.Errorf("Failed to patch RKE cluster yaml: %v", err)
	}

	return objYml, obj, nil
}

// patchRKEClusterYaml is needed due to auditv1.Policy{} doesn't provide yaml tags
func patchRKEClusterYaml(d *schema.ResourceData, in *rancher.RancherKubernetesEngineConfig) (string, error) {
	outFixed := make(map[string]interface{})
	if in.Services.KubeAPI.AuditLog != nil && in.Services.KubeAPI.AuditLog.Configuration != nil {
		inJSON, err := interfaceToJSON(in.Services.KubeAPI.AuditLog.Configuration.Policy)
		if err != nil {
			return "", err
		}
		if len(inJSON) > 0 {
			outFixed["audit_log"], err = jsonToMapInterface(inJSON)
			if err != nil {
				return "", fmt.Errorf("ummarshalling auditlog json: %s", err)
			}
		}
	}
	if in.Services.KubeAPI.EventRateLimit != nil && in.Services.KubeAPI.EventRateLimit.Configuration != nil {
		inJSON, err := interfaceToJSON(in.Services.KubeAPI.EventRateLimit.Configuration)
		if err != nil {
			return "", err
		}
		if len(inJSON) > 0 {
			outFixed["event_rate_limit"], err = jsonToMapInterface(inJSON)
			if err != nil {
				return "", fmt.Errorf("ummarshalling event_rate_limit json: %s", err)
			}
		}
	}
	if in.Services.KubeAPI.SecretsEncryptionConfig != nil && in.Services.KubeAPI.SecretsEncryptionConfig.CustomConfig != nil {
		customConfigV1Str, err := interfaceToGhodssyaml(in.Services.KubeAPI.SecretsEncryptionConfig.CustomConfig)
		if err != nil {
			return "", fmt.Errorf("Mashalling custom_config yaml: %v", err)
		}
		customConfigV1 := &apiserverconfigv1.EncryptionConfiguration{}
		err = ghodssyamlToInterface(customConfigV1Str, customConfigV1)
		if err != nil {
			return "", fmt.Errorf("Unmashalling custom_config yaml: %v", err)
		}
		inJSON, err := interfaceToJSON(customConfigV1)
		if err != nil {
			return "", err
		}
		if len(inJSON) > 0 {
			outFixed["secrets_encryption_config"], err = jsonToMapInterface(inJSON)
			if err != nil {
				return "", fmt.Errorf("ummarshalling eventrate json: %s", err)
			}
		}
	}

	outYml, err := interfaceToYaml(in)
	if err != nil {
		return "", fmt.Errorf("Failed to marshal yaml RKE cluster: %v", err)
	}

	if len(outFixed) == 0 {
		return outYml, nil
	}

	out := make(map[string]interface{})
	err = ghodssyamlToInterface(outYml, &out)
	if err != nil {
		return "", fmt.Errorf("ummarshalling RKE cluster yaml: %s", err)
	}

	if services, ok := out["services"].(map[string]interface{}); ok {
		if kubeapi, ok := services["kube-api"].(map[string]interface{}); ok {
			if auditlog, ok := kubeapi["audit_log"].(map[string]interface{}); ok && outFixed["audit_log"] != nil {
				if _, ok := auditlog["configuration"].(map[string]interface{}); ok {
					out["services"].(map[string]interface{})["kube-api"].(map[string]interface{})["audit_log"].(map[string]interface{})["configuration"].(map[string]interface{})["policy"] = outFixed["audit_log"]
				}
			}
			if _, ok := kubeapi["event_rate_limit"].(map[string]interface{}); ok && outFixed["event_rate_limit"] != nil {
				out["services"].(map[string]interface{})["kube-api"].(map[string]interface{})["event_rate_limit"].(map[string]interface{})["configuration"] = outFixed["event_rate_limit"]
			}
			if _, ok := kubeapi["secrets_encryption_config"].(map[string]interface{}); ok && outFixed["secrets_encryption_config"] != nil {
				secretEncryption := map[string]interface{}{}
				if dataServices, ok := d.Get("services").([]interface{}); ok && len(dataServices) > 0 {
					if secretEncryptionStr, ok := dataServices[0].(map[string]interface{})["kube_api"].([]interface{})[0].(map[string]interface{})["secrets_encryption_config"].([]interface{})[0].(map[string]interface{})["custom_config"].(string); ok {
						secretEncryption, _ = ghodssyamlToMapInterface(secretEncryptionStr)
					}
				}
				out["services"].(map[string]interface{})["kube-api"].(map[string]interface{})["secrets_encryption_config"].(map[string]interface{})["custom_config"] = secretEncryption
			}
		}
	}

	outYaml, err := interfaceToGhodssyaml(out)
	if err != nil {
		return "", fmt.Errorf("marshalling RKE cluster patched yaml: %s", err)
	}

	return outYaml, nil
}

func expandRKEClusterFlag(in *schema.ResourceData, clusterFilePath string) cluster.ExternalFlags {
	if in == nil {
		return cluster.ExternalFlags{}
	}

	updateOnly := in.Get("update_only").(bool)
	disablePortCheck := in.Get("disable_port_check").(bool)
	dind := in.Get("dind").(bool)

	if dind {
		updateOnly = false
	}

	// setting up the flags
	obj := cluster.GetExternalFlags(false, updateOnly, disablePortCheck, "", clusterFilePath)
	obj.DinD = dind
	if !dind {
		// Custom certificates and certificate dir flags
		if v, ok := in.Get("cert_dir").(string); ok && len(v) > 0 {
			obj.CertificateDir = v
		}
		obj.CustomCerts = in.Get("custom_certs").(bool)
	}

	return obj
}
