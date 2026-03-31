package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

func RegisterReadLoadBalancers(s *server.MCPServer, clientManager *oscclient.ClientManager) {
	tool := mcp.NewTool("osc_read_load_balancers",
		mcp.WithDescription(`List and inspect Load Balancers in your Outscale account.

Use this tool to:
- Check load balancer configurations
- View listeners (frontend ports and protocols)
- Inspect backend VMs and health checks
- Debug load balancing issues
- Find load balancers by name or Net ID`),
		mcp.WithString("load_balancer_names",
			mcp.Description("Filter by Load Balancer names (comma-separated)"),
		),
		mcp.WithString("profile",
			mcp.Description("Profile name to use (optional, uses default if not specified)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return withClient(ctx, clientManager, req, func(authCtx context.Context, client *oscclient.Client, profile string) (*mcp.CallToolResult, error) {
			return handleReadLoadBalancers(authCtx, client, req, profile)
		})
	})
}

func handleReadLoadBalancers(authCtx context.Context, client *oscclient.Client, req mcp.CallToolRequest, profile string) (*mcp.CallToolResult, error) {
	filters := osc.FiltersLoadBalancer{}
	args := req.Params.Arguments

	if lbNames := getString(args, "load_balancer_names"); lbNames != "" {
		filters.SetLoadBalancerNames(parseCommaSeparated(lbNames))
	}

	readReq := osc.ReadLoadBalancersRequest{}
	readReq.SetFilters(filters)

	read, _, err := client.API.LoadBalancerApi.ReadLoadBalancers(authCtx).ReadLoadBalancersRequest(readReq).Execute()
	if err != nil {
		return formatError("read load balancers", err), nil
	}

	lbs := make([]map[string]interface{}, 0)
	if read.LoadBalancers != nil {
		for _, lb := range *read.LoadBalancers {
			lbs = append(lbs, formatLoadBalancer(lb))
		}
	}

	response := map[string]interface{}{
		"load_balancers": lbs,
		"count":          len(lbs),
		"profile":        profile,
		"request_id":     safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}

func formatLoadBalancer(lb osc.LoadBalancer) map[string]interface{} {
	return map[string]interface{}{
		"name":                   safeString(lb.LoadBalancerName),
		"dns_name":               safeString(lb.DnsName),
		"type":                   safeString(lb.LoadBalancerType),
		"net_id":                 safeString(lb.NetId),
		"public_ip":              safeString(lb.PublicIp),
		"subnets":                safeStringSlice(lb.Subnets),
		"subregion_names":        safeStringSlice(lb.SubregionNames),
		"security_groups":        safeStringSlice(lb.SecurityGroups),
		"backend_vm_ids":         safeStringSlice(lb.BackendVmIds),
		"backend_ips":            safeStringSlice(lb.BackendIps),
		"secured_cookies":        safeBool(lb.SecuredCookies),
		"listeners":              formatListeners(lb.Listeners),
		"health_check":           formatHealthCheck(lb.HealthCheck),
		"sticky_cookie_policies": formatStickyCookiePolicies(lb.LoadBalancerStickyCookiePolicies),
		"app_sticky_policies":    formatAppStickyPolicies(lb.ApplicationStickyCookiePolicies),
		"access_log":             formatAccessLog(lb.AccessLog),
		"source_security_group":  formatSourceSecurityGroup(lb.SourceSecurityGroup),
		"tags":                   formatTags(lb.Tags),
	}
}

func formatListeners(listeners *[]osc.Listener) []map[string]interface{} {
	if listeners == nil {
		return []map[string]interface{}{}
	}
	result := make([]map[string]interface{}, len(*listeners))
	for i, l := range *listeners {
		result[i] = map[string]interface{}{
			"load_balancer_port":     safeInt32(l.LoadBalancerPort),
			"load_balancer_protocol": safeString(l.LoadBalancerProtocol),
			"backend_port":           safeInt32(l.BackendPort),
			"backend_protocol":       safeString(l.BackendProtocol),
			"server_certificate_id":  safeString(l.ServerCertificateId),
			"policy_names":           safeStringSlice(l.PolicyNames),
		}
	}
	return result
}

func formatHealthCheck(hc *osc.HealthCheck) map[string]interface{} {
	if hc == nil {
		return nil
	}
	return map[string]interface{}{
		"protocol":            hc.Protocol,
		"port":                hc.Port,
		"path":                safeString(hc.Path),
		"check_interval":      hc.CheckInterval,
		"timeout":             hc.Timeout,
		"healthy_threshold":   hc.HealthyThreshold,
		"unhealthy_threshold": hc.UnhealthyThreshold,
	}
}

func formatStickyCookiePolicies(policies *[]osc.LoadBalancerStickyCookiePolicy) []map[string]interface{} {
	if policies == nil {
		return []map[string]interface{}{}
	}
	result := make([]map[string]interface{}, len(*policies))
	for i, p := range *policies {
		result[i] = map[string]interface{}{
			"cookie_expiration_period": safeInt(p.CookieExpirationPeriod),
			"policy_name":              safeString(p.PolicyName),
		}
	}
	return result
}

func formatAppStickyPolicies(policies *[]osc.ApplicationStickyCookiePolicy) []map[string]interface{} {
	if policies == nil {
		return []map[string]interface{}{}
	}
	result := make([]map[string]interface{}, len(*policies))
	for i, p := range *policies {
		result[i] = map[string]interface{}{
			"cookie_name": safeString(p.CookieName),
			"policy_name": safeString(p.PolicyName),
		}
	}
	return result
}

func formatAccessLog(al *osc.AccessLog) map[string]interface{} {
	if al == nil {
		return nil
	}
	return map[string]interface{}{
		"is_enabled":           safeBool(al.IsEnabled),
		"osu_bucket_name":      safeString(al.OsuBucketName),
		"osu_bucket_prefix":    safeString(al.OsuBucketPrefix),
		"publication_interval": safeInt(al.PublicationInterval),
	}
}

func formatSourceSecurityGroup(ssg *osc.SourceSecurityGroup) map[string]interface{} {
	if ssg == nil {
		return nil
	}
	return map[string]interface{}{
		"name":       safeString(ssg.SecurityGroupName),
		"account_id": safeString(ssg.SecurityGroupAccountId),
	}
}

func safeStringSlice(s *[]string) []string {
	if s == nil {
		return []string{}
	}
	return *s
}

func formatTags(tags *[]osc.ResourceTag) []map[string]string {
	if tags == nil {
		return []map[string]string{}
	}
	result := make([]map[string]string, len(*tags))
	for i, t := range *tags {
		result[i] = map[string]string{
			"key":   t.Key,
			"value": t.Value,
		}
	}
	return result
}
