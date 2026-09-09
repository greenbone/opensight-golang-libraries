// SPDX-FileCopyrightText: 2026 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package computed holds the canonical names of the normalized `computed` bucket
// fields (the topology node-property contract): the keys discovery emits and the
// topology graph-build reads. Using these constants instead of string literals
// makes a misspelled or divergent key (parentID vs parentId, vpcId vs networkId,
// account_id vs accountId) a compile error instead of a silently broken graph
// join. Raw provider-specific keys are NOT here: they are not the contract, and
// a key earns a constant only once both sides agree on it.
package computed

// Common envelope (every node).
const (
	SourceProvider = "sourceProvider"
	ResourceType   = "resourceType"
	ResourceID     = "resourceId"
	AccountID      = "accountId"
	AccountName    = "accountName"
	OrganizationID = "organizationId"
	Region         = "region"
	Zone           = "zone"
)

// Tenancy / Governance.
const (
	ParentID = "parentId" // Contains FK: child -> parent node id
	// ParentVisibility distinguishes "no parent" from "parent not visible to
	// the scanning principal": "unknown" means the real parent exists but was
	// invisible, so producers leave ParentID empty instead of inventing a
	// tenant-direct edge, and coverage/reachability can see the partial
	// hierarchy state.
	ParentVisibility    = "parentVisibility"
	State               = "state"
	OrgKind             = "orgKind"     // aws_organization | entra_tenant | gcp_organization
	AccountKind         = "accountKind" // aws_account | azure_subscription | gcp_project
	AltIDs              = "altIds"
	ManagementAccountID = "managementAccountId"
	ARN                 = "arn"
)

// GCP-native aliases
const (
	ProjectID   = "projectId"
	ProjectName = "projectName"
)

// Network cores
const (
	NetworkID           = "networkId"
	CidrBlocks          = "cidrBlocks"
	IPv6CidrBlocks      = "ipv6CidrBlocks"
	IsDefault           = "isDefault"
	DNSServers          = "dnsServers"
	DdosProtected       = "ddosProtected"
	RoutingMode         = "routingMode"
	RouteTableID        = "routeTableId"
	SecurityGroupIDs    = "securityGroupIds"
	NatGatewayID        = "natGatewayId"
	AutoAssignPublicIP  = "autoAssignPublicIp"
	PrivateGoogleAccess = "privateGoogleAccess"
	Purpose             = "purpose"
	Kind                = "kind" // RouteTable: network | transit_hub (R10)
	Routes              = "routes"
	Rules               = "rules"
	AssociatedSubnetIDs = "associatedSubnetIds"
	IsMain              = "isMain"
	SubnetID            = "subnetId" // InSubnet FK: NIC -> Subnet node id
	AttachedInstanceID  = "attachedInstanceId"
	PrivateIPs          = "privateIps"
	PublicIPIDs         = "publicIpIds"
	IPForwardingEnabled = "ipForwardingEnabled"
	AttachedNicIDs      = "attachedNicIds"
	AttachedSubnetIDs   = "attachedSubnetIds"
	// ApplicationSecurityGroupIDs is a NIC's Azure ASG membership: the set of
	// applicationSecurityGroups its ipConfigurations belong to. It is what lets an
	// NSG rule's asg: ref resolve to member NICs
	ApplicationSecurityGroupIDs = "applicationSecurityGroupIds"
	TargetTags                  = "targetTags"
	TargetServiceAccounts       = "targetServiceAccounts"
	Scope                       = "scope" // NSG: org | folder (GCP hierarchical firewall policy)
	PrefixListID                = "prefixListId"
	Cidrs                       = "cidrs"
	AddressFamily               = "addressFamily"
	Owner                       = "owner"
)

// Gateways: NAT / CloudRouter / Bastion / ManagedFirewall / FirewallPolicy.
const (
	SubnetIDs           = "subnetIds"           // InSubnet FK (plural): NatGateway/ManagedFirewall -> Subnet node ids
	ConnectivityType    = "connectivityType"    // NatGateway: public | private
	ASN                 = "asn"                 // CloudRouter BGP ASN
	HasCloudNat         = "hasCloudNat"         // CloudRouter carries >=1 NAT config (NAT synthesis anchor)
	SKU                 = "sku"                 // BastionHost SKU tier (Basic | Standard | Developer | Premium)
	TunnelingEnabled    = "tunnelingEnabled"    // BastionHost native-client tunneling (widens pivot beyond 22/3389)
	IPConnectEnabled    = "ipConnectEnabled"    // BastionHost connect-by-IP (widens fan-out targets)
	ScaleUnits          = "scaleUnits"          // BastionHost throughput units
	PolicyID            = "policyId"            // ManagedFirewall -> FirewallPolicy (rule source)
	AttachedFirewallIDs = "attachedFirewallIds" // FirewallPolicy -> ManagedFirewall (Enforces)
	DefaultAction       = "defaultAction"       // FirewallPolicy fail-open vs fail-closed default
	FlowLogsEnabled     = "flowLogsEnabled"     // VirtualNetwork: NetworkWatcher flow logs present (feeder fold)
	// AssociatedNetworkIDs is the AppliesTo FK list for a resource bound to one or
	// more VirtualNetworks rather than a subnet: a GCP network firewall policy
	// (associations[].attachmentTarget) or a firewall endpoint association.
	AssociatedNetworkIDs = "associatedNetworkIds"
	// AttachedScopeIds is the AppliesTo FK list for a GCP HIERARCHY firewall policy:
	// the Org/Folder node ids it is ACTUALLY associated with (associations[].attachmentTarget),
	// which may be zero, one, or several and is distinct from the policy's creation parent.
	AttachedScopeIds = "attachedScopeIds"
)

// IP addressing + DNS : PublicIpAddress / DnsZone / DnsRecord.
const (
	// IPAddress is the literal public address a PublicIpAddress carries; the join
	// key a DnsRecord's values resolve onto (ResolvesTo).
	IPAddress = "ipAddress"
	// Allocation is static | ephemeral: a reserved/EIP address vs an
	// auto-assigned one (dangling-DNS and idle-IP CSPM).
	Allocation = "allocation"
	// AttachedResourceID is the NIC / NAT / Bastion / VPN gateway a public IP
	// fronts: the Exposes FK (empty when the IP is idle).
	AttachedResourceID = "attachedResourceId"
	// Fqdn is a public IP's provider DNS name or a DNS record's fully-qualified
	// name (takeover / enumeration).
	Fqdn = "fqdn"
	// DomainName is a DnsZone's apex domain (the record-join anchor).
	DomainName = "domainName"
	// Visibility is public | private for a DnsZone (entry-surface split).
	Visibility = "visibility"
	// LinkedNetworkIDs is a private DnsZone's LinkedTo FK list: the VNets it
	// resolves in.
	LinkedNetworkIDs = "linkedNetworkIds"
	// ZoneID is a DnsRecord's Contains FK: its parent DnsZone node id.
	ZoneID = "zoneId"
	// RecordType is a DnsRecord's type (A | AAAA | CNAME | ALIAS | ...).
	RecordType = "recordType"
	// TTL is a DnsRecord's time-to-live in seconds.
	TTL = "ttl"
	// Values is a DnsRecord's right-hand side (rrdatas / resource records): the
	// ResolvesTo join values.
	Values = "values"
	// AliasTargetID is an alias/apex record's target (the AWS alias DNSName,
	// Azure targetResource.id): a second ResolvesTo path (PublicIp | LB | Cdn).
	AliasTargetID = "aliasTargetId"
)

// Cross-network connectivity: peering / transit hub / private link /
// VPN. Peering is NOT transitive; the transitivity-gate fields are collected so
// the segment-adjacency pass can decide, never assume a mesh.
const (
	// VpcPeering.
	LocalNetworkID        = "localNetworkId"  // PeersWith: the owning VirtualNetwork
	RemoteNetworkID       = "remoteNetworkId" // PeersWith: the peer VirtualNetwork (may be :External)
	RemoteAccountID       = "remoteAccountId" // crossAccount flag + :External synth
	RemoteRegion          = "remoteRegion"
	RemoteCidrBlocks      = "remoteCidrBlocks"      // undiscovered-side reachability (peering + vpn)
	AllowForwardedTraffic = "allowForwardedTraffic" // transitivity gate
	AllowGatewayTransit   = "allowGatewayTransit"   // transitivity gate
	UseRemoteGateways     = "useRemoteGateways"     // transitivity gate
	// TransitHub.
	DefaultRouteTableAssociation = "defaultRouteTableAssociation"
	AddressPrefix                = "addressPrefix"
	// HubAttachments is the transit-hub attachment list ({resourceId,
	// associatedRouteTableId, propagatedRouteTableIds}) the AttachedToHub edge and
	// R6 ("hub route tables decide, never assume mesh") read.
	HubAttachments = "hubAttachments"
	// VpcEndpoint / PrivateEndpoint / PSC consumer.
	EndpointKind      = "endpointKind"    // gateway | interface | psc
	NicIDs            = "nicIds"          // the endpoint's data-path NICs
	TargetServiceID   = "targetServiceId" // PrivateLinkTo: the producer service
	PrivateDNSEnabled = "privateDnsEnabled"
	// VpcEndpointService / PrivateLinkService / ServiceAttachment (producer).
	BackendLbID        = "backendLbId"        // Backs: the LB behind the service
	AcceptanceRequired = "acceptanceRequired" // toxic: open producer when false
	AllowedConsumers   = "allowedConsumers"   // cross-tenant exposure
	// VpnGateway / VpnConnection.
	PublicIps        = "publicIps"    // VpnGateway -> Exposes
	GatewayID        = "gatewayId"    // ConnectsTo: VpnConnection -> VpnGateway
	RemoteSiteID     = "remoteSiteId" // ConnectsTo: the on-prem site
	StaticRoutesOnly = "staticRoutesOnly"
	TunnelCount      = "tunnelCount"
	// Hybrid: DedicatedCircuit + ClientVpnEndpoint. OnPremSite reuses
	// publicIps/cidrBlocks/asn.
	Bandwidth              = "bandwidth"
	RemoteAsn              = "remoteAsn"              // DedicatedCircuit BGP peer ASN
	AttachedHubOrGatewayID = "attachedHubOrGatewayId" // DedicatedCircuit -> AttachedToHub / ConnectsTo
	ClientCidrBlocks       = "clientCidrBlocks"       // ClientVpnEndpoint remote-client address pool
	AuthType               = "authType"               // ClientVpnEndpoint (cert | directory | federated)
	SplitTunnel            = "splitTunnel"            // ClientVpnEndpoint exfil-path signal
	AuthorizationRules     = "authorizationRules"     // ClientVpnEndpoint reachability gate (group -> target CIDR)
)

// Ingress: the internet -> workload funnel. LoadBalancer /
// ApplicationGateway reuse networkId / subnetIds / publicIpIds / securityGroupIds;
// LbListener / Cdn reuse the shared object-array serialization (jsonArrayKeys on
// the consumer side) for their action/rule arrays.
const (
	// LoadBalancer / ApplicationGateway.
	Scheme      = "scheme"      // internet-facing | internal (attack-entry class)
	Layer       = "layer"       // l4 | l7 (rule targeting)
	DNSName     = "dnsName"     // LB/GAE/TrafficManager DNS name (ResolvesTo join)
	WafPolicyID = "wafPolicyId" // reverse of Protects
	CdnEnabled  = "cdnEnabled"  // GCP Cloud CDN is a flag on the LB, not a node
	// LbListener.
	LbID           = "lbId" // HasListener: listener -> its load balancer
	Port           = "port"
	Protocol       = "protocol"
	CertificateIDs = "certificateIds" // TLS posture (CSPM)
	DefaultActions = "defaultActions" // ForwardsTo: {type, backendPoolId} per action
	// LbBackendPool.
	TargetKind = "targetKind" // instance | ip | lambda | neg
	Targets    = "targets"    // Targets: {id|ip, port}; unresolved -> :External
	// GlobalAcceleratorEndpoint (AWS) / TrafficManager (Azure).
	StaticIps       = "staticIps"       // GAE anycast IPs (ResolvesTo)
	Enabled         = "enabled"         // live edge gate
	ListenerPorts   = "listenerPorts"   // GAE entry surface
	EndpointTargets = "endpointTargets" // GAE Targets: {id, weight, clientIpPreserved}
	RoutingMethod   = "routingMethod"   // TrafficManager routing analysis
	Endpoints       = "endpoints"       // TrafficManager ResolvesTo: {targetId|target, kind, ...}
	MonitorConfig   = "monitorConfig"   // TrafficManager health probe (CSPM)
	// ApiGateway.
	Exposure         = "exposure"         // public | private (attack-entry class)
	VpcLinkTargetIDs = "vpcLinkTargetIds" // RoutesTraffic
	CustomDomains    = "customDomains"
	AuthorizerCount  = "authorizerCount" // CSPM
	// Cdn.
	DomainNames     = "domainNames"
	OriginHosts     = "originHosts" // FrontsFor: LB | ObjectStorage | External
	OriginIDs       = "originIds"
	ViewerTLSPolicy = "viewerTlsPolicy" // CSPM
	// WafPolicy.
	Mode                = "mode"                // block | detect (CSPM)
	ManagedRuleSets     = "managedRuleSets"     // preconfigured WAF rules (CSPM)
	AttachedResourceIDs = "attachedResourceIds" // Protects: LB | AppGw | ApiGw | Cdn
)

// Data-plane gates (leaf/compute assets: DB, Storage, Cache, k8s). These
// are producer-emitted normalized fields that reachability reads as a
// SECOND ingress gate, independent of any NSG: a public endpoint is only reachable
// when its own allow-list admits the peer. They are registered here AHEAD of the
// clients that emit them so the discovery bucket split routes them to Computed the moment
// a client emits one; an unregistered key would fall to Properties and be invisible
// to reachability (a false-negative public exposure).
const (
	// PublicNetworkAccess is the data-plane public gate (enabled | disabled |
	// restricted): RDS PubliclyAccessible, Azure publicNetworkAccess, GCP
	// publicAccessPrevention / ipConfiguration.ipv4Enabled, normalized.
	PublicNetworkAccess = "publicNetworkAccess"
	// AuthorizedCidrs is the flat allow-list a public endpoint admits, resolved
	// from SG ingress (AWS), firewall ipRules (Azure), or authorizedNetworks (GCP).
	AuthorizedCidrs = "authorizedCidrs"
	// APIServerPublic is the KubernetesCluster control-plane public-endpoint flag
	// (EKS endpointPublicAccess, AKS !enablePrivateCluster, GKE !privateEndpoint).
	APIServerPublic = "apiServerPublic"
	// APIServerAuthorizedCidrs is the k8s control-plane allow-list (EKS
	// publicAccessCidrs, AKS authorizedIPRanges, GKE masterAuthorizedNetworks).
	APIServerAuthorizedCidrs = "apiServerAuthorizedCidrs"
	// PrivateNodes marks a KubernetesCluster whose nodes have no public IPs.
	PrivateNodes = "privateNodes"
)

// Compute-instance fields (VirtualMachine and the workload leaves that
// carry an OS / identity / metadata surface). Network placement rides the shared
// nicIds/subnetIds/securityGroupIds keys; these are the instance-plane fields the
// graph and downstream (vuln binding, IAM escalation, CSPM) read.
const (
	// PowerState is the normalized run state (running | stopped | terminated |
	// unknown): only a running VM is an attack node.
	PowerState = "powerState"
	// OSType is linux | windows (AWS PlatformDetails, Azure osDisk.osType, GCP
	// disk license), used for vulnerability binding.
	OSType = "osType"
	// ImageID is the base image the instance booted (AWS ImageId, Azure
	// imageReference, GCP sourceImage), for CVE binding.
	ImageID = "imageId"
	// IdentityIDs are the workload identities attached to the instance (AWS
	// instance-profile ARN, Azure principal + userAssigned, GCP service-account
	// emails): the VM -> IAM escalation link.
	IdentityIDs = "identityIds"
	// IdentityScopes are the OAuth scopes a GCP service account grants the
	// instance (escalation breadth).
	IdentityScopes = "identityScopes"
	// NetworkTags are the GCP instance network tags, the join key GCP firewall /
	// route rules target (targetTags).
	NetworkTags = "networkTags"
	// MetadataService captures the instance metadata posture (AWS IMDSv2
	// requirement, GCP metadata items) for the IMDS/SSRF CSPM check.
	MetadataService = "metadataService"
	// InstanceType is the node machine size (AWS EKS Nodegroup InstanceTypes,
	// Azure AKS agentPool vmSize, GCP GKE NodePool config.machineType), for UI
	// and vulnerability sizing context.
	InstanceType = "instanceType"
	// ImageType is the node OS image family (AWS EKS AmiType, Azure AKS
	// osSKU, GCP GKE NodePool config.imageType), for vulnerability binding.
	ImageType = "imageType"
	// LaunchTemplateID is the launch spec an AutoScalingGroup scales from (AWS
	// ASG LaunchTemplate/LaunchConfiguration, GCP MIG instanceTemplate); the
	// Uses edge to the LaunchTemplate node.
	LaunchTemplateID = "launchTemplateId"
	// BackendPoolIDs are the load-balancer backend pools an AutoScalingGroup or
	// scale set registers into (the scale-out inherits the LB's exposure).
	BackendPoolIDs = "backendPoolIds"
	// Ports is a workload's listening port set (ECS task / target-group ports,
	// ACI ipAddress.ports), the reachability port class.
	Ports = "ports"
	// NetworkOrigin is an S3 access point's reachability scope (VPC | Internet):
	// a VPC origin is reachable only from inside its network.
	NetworkOrigin = "networkOrigin"
	// BucketID is the ObjectStorage an S3 access point fronts (its FrontsFor
	// target).
	BucketID = "bucketId"
)

// Serverless fields (ServerlessFunction: Lambda / Functions / Cloud Run,
// Cloud Functions). Network placement rides subnetIds/securityGroupIds; these are
// the function-plane fields the graph reads to decide direct internet entry.
const (
	// PublicURL is the function's HTTPS invoke endpoint (Lambda function URL, the
	// Functions/Cloud Run default hostname), the direct-entry surface.
	PublicURL = "publicUrl"
	// UrlAuthType gates the PublicURL (Lambda NONE | AWS_IAM; Cloud Run
	// allUsers vs authenticated): NONE / public = unauthenticated direct entry.
	UrlAuthType = "urlAuthType"
	// Ingress is the Cloud Run / Functions ingress setting (all | internal |
	// internal-and-cloud-load-balancing), the network-side entry gate.
	Ingress = "ingress"
	// Runtime is the function runtime (Lambda Runtime, Functions runtime, Cloud
	// Run image), the CVE-binding coordinate (the serverless analogue of imageId).
	Runtime = "runtime"
)

// IAM plane, principals - the AWS vertical slice's keys; the group grows the
// same way the network groups above did, as more providers land.
const (
	// IdentitySource discriminates sub-kinds under one canonical principal
	// label (aws_iam_user | aws_sso_user | aws_iam_group | aws_sso_group | ...).
	IdentitySource = "identitySource"
	// GroupIds is the MemberOf edge source: ids of the groups this principal
	// belongs to.
	GroupIds = "groupIds"
	// MfaPresent rolls up whether the user has at least one MFA device.
	MfaPresent = "mfaPresent"
	// AttachedPolicyIds is the AttachedPolicy edge source (native IAM only;
	// an SSO principal's access path is AccountAssignments instead).
	AttachedPolicyIds = "attachedPolicyIds"
	// AccountAssignments is an object array [{permissionSetArn, accountId}]
	// (SSO principals only), stored as JSON at graph ingest.
	AccountAssignments = "accountAssignments"
	// PermissionsBoundaryPolicyID is the BoundedBy ceiling edge source.
	PermissionsBoundaryPolicyID = "permissionsBoundaryPolicyId"
	// PrincipalKind discriminates workload-principal sub-kinds under
	// ServicePrincipal (AWS role has no sub-kind; azure/gcp set theirs).
	PrincipalKind = "principalKind"
	// MaxSessionDuration is the role's session ceiling in seconds.
	MaxSessionDuration = "maxSessionDuration"
	// TrustedPrincipals is an object array of parsed trust-policy statements
	// (the CanAssume edge source), stored as JSON at graph ingest.
	TrustedPrincipals = "trustedPrincipals"
	// TrustedServices is an object array of service-principal trust entries
	// (closed-set operational plumbing, no edge), stored as JSON at graph ingest.
	TrustedServices = "trustedServices"
	// Email is a principal's email (GCP ServiceAccount/UserIdentity/IamGroup):
	// the join key GCP IAM bindings reference members by (user:/serviceAccount:
	// {email}), never by the node's own canonical id.
	Email = "email"
	// Disabled is the live/dead gate on a GCP ServiceAccount (ServiceAccount.Disabled).
	Disabled = "disabled"
	// IamBindings is an object array of GCP IAM policy bindings on a
	// GovernanceBinding sidecar node (the HasBinding edge source), stored as JSON
	// at graph ingest.
	IamBindings = "iamBindings"
	// AttachmentPointID is the AppliesTo FK on a GCP GovernanceBinding sidecar:
	// the id of the tenancy node (or ServicePrincipal, for the sa_self case) the
	// bindings are attached to. A distinct key, not ResourceID, for the same
	// reason as PolicyResourceID.
	AttachmentPointID = "attachmentPointId"
)

// IAM plane, resource-attached bindings - the AWS resource-policy
// vertical slice's keys, on the ResourcePolicy sidecar node.
const (
	// ResourcePolicyGrants is an object array of parsed resource-based-policy
	// statements (the GrantedAccess edge source), stored as JSON at graph ingest.
	ResourcePolicyGrants = "resourcePolicyGrants"
	// GrantedServices is an object array of service-principal grants from a
	// resource-based policy (closed-set, no edge), stored as JSON at graph ingest.
	GrantedServices = "grantedServices"
	// PolicyResourceID is the AppliesTo FK: the id of the resource a
	// ResourcePolicy sidecar node is attached to. It is a distinct key, not
	// ResourceID, because ResourceID is the node's OWN id ("{resourceId}#policy")
	// and doubles as the resourceId property in the graph build - reusing it for
	// the target would make AppliesTo match the policy node against itself.
	PolicyResourceID = "policyResourceId"
	// RoleID is the ResolvesTo FK on an AWS InstanceProfile node: the ARN of the
	// role currently attached to the profile (InstanceProfile -> ServicePrincipal),
	// the second hop of the compute-attached-identity join.
	RoleID = "roleId"
	// FederationType discriminates a SamlProvider/OidcProvider federation node:
	// "saml" | "oidc" (AWS/Azure external IdP) | "workload_identity" |
	// "workload_identity_pool" | "workforce_identity_pool".
	FederationType = "federationType"
	// Issuer is the federation trust anchor's issuer - who may present a token:
	// an OIDC provider Url, a SAML metadata entityID, or a pool provider's issuer.
	Issuer = "issuer"
	// ClientIDList is the audience allow-list on an AWS OidcProvider node
	// (GetOpenIDConnectProvider ClientIDList) - the tokens' accepted aud values.
	ClientIDList = "clientIdList"
	// ThumbprintList is the server-cert thumbprint set on an AWS OidcProvider node
	// (GetOpenIDConnectProvider ThumbprintList).
	ThumbprintList = "thumbprintList"
	// PoolKind discriminates a federation pool/provider node's platform:
	// "gcp_workload_identity_pool" | "gcp_workforce_identity_pool" |
	// "aws_cognito_identity_pool".
	PoolKind = "poolKind"
	// PoolID is the BelongsToPool FK on a GCP pool-provider node (SamlProvider/
	// OidcProvider, poolKind gcp_workload_identity_pool): the id of the parent
	// WorkloadIdentityPool the provider belongs to.
	PoolID = "poolId"
	// SubjectCondition is the federation trust's subject scope - the actual
	// security boundary: a GCP pool provider's attributeCondition, an Azure FIC
	// subject, the "who exactly may present a token" half of the trust. An empty
	// value is a meaningfully broad (unscoped) trust, not an absent one.
	SubjectCondition = "subjectCondition"
	// AttributeMapping is a GCP WorkloadIdentityPoolProvider's claim->attribute map
	// (e.g. google.subject <- assertion.sub), an object of string mappings.
	AttributeMapping = "attributeMapping"
	// AllowUnauthenticatedIdentities is true on an AWS Cognito Identity Pool
	// (WorkloadIdentityPool, poolKind aws_cognito_identity_pool) that lets an
	// anonymous caller assume an IAM role with no credential at all - a direct
	// public reachability source.
	AllowUnauthenticatedIdentities = "allowUnauthenticatedIdentities"
	// RoleTargets is an object array on an AWS Cognito Identity Pool: the roles the
	// pool federates identities into, one entry per default (authenticated/
	// unauthenticated) mapping plus one per rules-based mapping. Each entry:
	// {mappingId, roleArn, requiresAuth, mappingType, provenance}.
	RoleTargets = "roleTargets"
	// HasTokenBasedRoleMapping is true on an AWS Cognito Identity Pool that selects
	// the role from the caller's own token claims at credential time (RoleMapping
	// Type Token) - a caller-controlled target the collector cannot resolve ahead
	// of time, flagged rather than silently dropped.
	HasTokenBasedRoleMapping = "hasTokenBasedRoleMapping"

	// Azure IAM. Tenant-composite ids are {tenantId}:{objId} (README
	// §2A1), so principal/role FKs below are collector-constructed in that form.
	// TenantID is the Azure tenant a node belongs to (the composite-id prefix).
	TenantID = "tenantId"
	// PrincipalID is the FK to the granted/assigned principal on an Azure
	// RoleAssignment / RoleEligibilitySchedule node: {tenantId}:{PrincipalID}, so it
	// joins to the Graph-collected principal node (or an UncollectedPrincipal stub).
	PrincipalID = "principalId"
	// RoleDefinitionID is the FK to the RoleDefinition on an Azure RoleAssignment /
	// RoleEligibilitySchedule node: {tenantId}:{lower(RoleDefinitionId)}.
	RoleDefinitionID = "roleDefinitionId"
	// ConditionJSON is a verbatim ABAC condition expression on an Azure
	// RoleAssignment node (RoleAssignment.Condition) - a single DSL string.
	ConditionJSON = "conditionJson"
	// ScheduleStatus is the PIM status on an Azure RoleEligibilitySchedule node.
	ScheduleStatus = "scheduleStatus"
	// RoleName is the human role name on an Azure RoleDefinition node.
	RoleName = "roleName"
	// Actions/NotActions/DataActions/NotDataActions are the Azure RoleDefinition
	// permission sets (string arrays) - the management- and data-plane action globs.
	Actions        = "actions"
	NotActions     = "notActions"
	DataActions    = "dataActions"
	NotDataActions = "notDataActions"
	// ArmIdentityID is the underlying armmsi.Identity ARM id on a user-assigned Azure
	// ManagedIdentity node - the secondary join key IAM-8 uses when a VM's
	// identityIds entry is an ARM id rather than a Graph object id (README §2b).
	ArmIdentityID = "armIdentityId"
	// AppID is the Azure application (client) id: on a ServicePrincipal it matches
	// the Application it was instantiated from (Application.appId, InstantiatedFrom).
	AppID = "appId"
	// AppOwnerOrganizationID is the tenant that owns the Application a multi-tenant
	// ServicePrincipal was instantiated from (may differ from the SP's own tenant).
	AppOwnerOrganizationID = "appOwnerOrganizationId"
	// CredentialCount / HasActiveCredential summarize an Azure ServicePrincipal's
	// password/key credentials (count + an unexpired-implies-active check).
	CredentialCount     = "credentialCount"     //nolint:gosec // G101 false positive: a property-name constant, not a credential
	HasActiveCredential = "hasActiveCredential" //nolint:gosec // G101 false positive: a property-name constant, not a credential
	// Audiences is the accepted-audience list on an Azure federated-identity-credential
	// OidcProvider node (the aud values an external token may carry).
	Audiences = "audiences"
	// OwnerApplicationID is the Trusts FK on an Azure workload-federation OidcProvider
	// (a Graph federatedIdentityCredential): {tenantId}:{Application object id}.
	OwnerApplicationID = "ownerApplicationId"
	// OwnerArmIdentityID is the Trusts FK on an Azure OidcProvider backed by a UAMI
	// federatedIdentityCredential: the underlying armmsi.Identity's ARM id.
	OwnerArmIdentityID = "ownerArmIdentityId"
	// OwnerOrganizationID is the Trusts FK on an Azure external-IdP SamlProvider/
	// OidcProvider (Graph identityProviders): the tenant that configured the trust.
	OwnerOrganizationID = "ownerOrganizationId"
	// IdentityProviderKind discriminates an IdentityProvider node's platform, e.g.
	// "aws_cognito_user_pool".
	IdentityProviderKind = "identityProviderKind"
	// AttachedIdpIds is the Trusts FK list on an AWS Cognito User Pool IdentityProvider:
	// the ids of its own IdpConfig children ({poolArn}#idp:{ProviderName}).
	AttachedIdpIds = "attachedIdpIds"
	// IdentityPoolTrustRefs is an AWS Cognito Identity Pool's own directly-trusted
	// external IdPs (independent of any user pool's AttachedIdpIds): an object array
	// [{kind, ref}] where ref is a real node id (user pool / SAML / OIDC provider ARN).
	// The Trusts FK for the identity-pool federation case. JSON-serialized at ingest.
	IdentityPoolTrustRefs = "identityPoolTrustRefs"
	// ProviderType is an AWS IdpConfig's IdP type (SAML|OIDC|Google|Facebook|
	// LoginWithAmazon|...); only SAML/OIDC carry a meaningful issuer.
	ProviderType = "providerType"

	// AppRoles is a resource Azure ServicePrincipal's published app-role catalog
	// ([{id,value,displayName}]), the id->permission mapping IAM-7's GrantedAppRole
	// resolves appRoleId against. Object array, JSON-serialized at ingest.
	AppRoles = "appRoles"
	// AssigneeID is an Azure AppRoleAssignment's granted principal, composite
	// {tenantId}:{Graph principalId} (a ServicePrincipal, UserIdentity, or IamGroup).
	AssigneeID = "assigneeId"
	// AssigneeType is the Graph principalType of the assignee: ServicePrincipal|User|
	// Group; GrantedAppRole routes to the matching node label by this.
	AssigneeType = "assigneeType"
	// ResourceSpID is an Azure AppRoleAssignment's resource ServicePrincipal, composite
	// {tenantId}:{Graph resourceId} - the SP whose API surface is being granted.
	ResourceSpID = "resourceSpId"
	// AppRoleID is the GUID of the granted app role on the resource SP (zero GUID means
	// default access, no API permission); GrantedAppRole resolves it via AppRoles.
	AppRoleID = "appRoleId"

	// RoleType discriminates an Azure RoleAssignment/RoleDefinition's plane: "resource"
	// (ARM RBAC, armauthorization - scope is an ARM resource path) vs "directory" (Entra
	// directory roles, Graph roleManagement/directory - Global Administrator et al., scope
	// is the tenant/AU). Same node types + edges (HasRoleAssignment/AppliesTo/AppliesAt),
	// different plane.
	RoleType = "roleType"

	// PolicyStatement node fields (Tier C effective-permission). One parsed statement of an
	// AWS identity/boundary/SCP/RCP policy - the evaluable inputs the ordered AWS policy
	// evaluation composes. Actions/NotActions (above) are reused for the action globs.
	//
	// Effect is the statement effect: "Allow" or "Deny".
	Effect = "effect"
	// Resources / NotResources are the statement's resource-ARN globs.
	Resources    = "resources"
	NotResources = "notResources"
	// IsNotAction / IsNotResource flag NotAction / NotResource statements (an inverted match
	// set, a frequent over-grant source, surfaced rather than silently expanded).
	IsNotAction   = "isNotAction"
	IsNotResource = "isNotResource"
	// ConditionParsed reports whether the statement's Condition block was fully evaluated
	// (false = a condition is present but unresolved, so a dependent decision is conditional,
	// never a definitive allow). ConditionJSON (above) holds the verbatim condition.
	ConditionParsed = "conditionParsed"
	// PolicyRole is the statement's evaluation role: "identity" (union with resource), "boundary"
	// (intersection ceiling), "scp" (org identity-side ceiling), "rcp" (org resource-side ceiling).
	PolicyRole = "policyRole"
	// SourcePolicyID is the policy the statement came from (managed-policy ARN or
	// {principalArn}#inline:{name}) - provenance for the evaluator's "via" trail.
	SourcePolicyID = "sourcePolicyId"
	// StatementSid is the statement's Sid (or a synthetic index id when absent).
	StatementSid = "statementSid"

	// AttachedTargets is the org-scope ids an AWS SCP/RCP PolicyStatement attaches to (root
	// r-xxx / OU ou-xxx -> OrganizationalUnit node, account 12-digit -> Account node) - the
	// AppliesTo FKs. A native string list (the target ids are node ids directly).
	AttachedTargets = "attachedTargets"

	// IncludedPermissions is a GCP custom-role's exact permission list (roles.get FULL view) -
	// a native string list, no globs (GCP custom roles enumerate permissions). Feeds the
	// custom-role escalation match (a bound custom role whose permissions include a privesc
	// permission like resourcemanager.projects.setIamPolicy).
	IncludedPermissions = "includedPermissions"

	// DenyRules is the object array of a GCP DenyPolicy's rules (Tier C deny input), one per
	// denyRule: {deniedPrincipals[], deniedPermissions[] (FQDN service.googleapis.com/
	// resource.verb), exceptionPrincipals[], condition, conditionParsed}. Serialized to JSON
	// at graph ingest (an object array, like iamBindings). The evaluator UNWINDs it and
	// checks it before any allow binding (GCP evaluates deny first).
	DenyRules = "denyRules"

	// Secrets & key plane. Graph-essential fields only; the fuller CSPM
	// property surface (rotation, expiry, key spec) is deferred to a later pass.
	//
	// EncryptionKeyId is the customer-managed key id/ARN a resource is encrypted with:
	// the EncryptedWith FK a data/secret/backup node emits so the graph can join it to
	// its EncryptionKey node (key blast radius: "this key decrypts these N resources").
	EncryptionKeyId = "encryptionKeyId"
	// EnableRbacAuthorization is an Azure Key Vault's authorization mode: true = ARM RBAC
	// (role assignments resolve via containment), false = the legacy access-policy model
	// whose grants live in AccessPolicies instead. Gates which GrantedAccess path applies.
	EnableRbacAuthorization = "enableRbacAuthorization"
	// AccessPolicies is an Azure access-policy-mode vault's inline grants (object array,
	// [{tenantId, objectId, permissions}]), the GrantedAccess edge source when
	// EnableRbacAuthorization is false. Serialized to JSON at graph ingest (like iamBindings).
	AccessPolicies = "accessPolicies"
	// SourceResourceId is the id of the resource a copy-type node derives from: the
	// SnapshotOf FK a snapshot/image emits so the graph can join it to its source
	// volume/DB/image node (a snapshot inherits its source's sensitivity). Same id
	// form as the source node's id.
	SourceResourceId = "sourceResourceId"
	// Encrypted is whether a copy-type node (snapshot/image) is encrypted at rest.
	// A CMEK copy also emits EncryptionKeyId; a platform-key copy sets this true with
	// no key node. Distinct from EncryptionKeyId so an unencrypted copy is explicit.
	Encrypted = "encrypted"
	// SharedPublicly is the producer's raw share-list verdict for a copy-type node: a
	// snapshot/image whose share list is `all` (or crosses accounts) is a full data
	// copy leak. The collector emits this fact; the exposure layer derives the
	// canonical isPublic flag (publicSource = snapshot_share) from it, never the
	// collector, so it does not collide with the exposure-owned isPublic reset.
	SharedPublicly = "sharedPublicly"
	// TargetBucketID is the object-storage bucket an audit trail writes its logs to:
	// the WritesTo FK an AuditTrail emits so the graph can join it to the bucket node
	// (the tamper path "who can write the bucket that stores the trail"). Same id form
	// as the ObjectStorage node's id.
	TargetBucketID = "targetBucketId"
	// MultiRegion is whether an audit trail records events in every region (a
	// single-region trail leaves other regions unlogged).
	MultiRegion = "multiRegion"
	// IsOrgTrail is whether an audit trail is an organization trail (covering every
	// account) rather than a single-account trail.
	IsOrgTrail = "isOrgTrail"
	// LoggingEnabled is whether an audit trail is actively logging (a stopped trail
	// records nothing even though it still exists).
	LoggingEnabled = "loggingEnabled"
	// LogFileValidationEnabled is whether an audit trail signs its log files for
	// tamper-evidence (its absence lets a log be altered undetectably).
	LogFileValidationEnabled = "logFileValidationEnabled"
	// RetentionDays is a log store's retention window in days (0 = never expires);
	// a short window shrinks the forensic record.
	RetentionDays = "retentionDays"
	// NotAfter is a certificate's expiry timestamp (RFC3339); an expired or
	// soon-expiring certificate is an availability and trust risk.
	NotAfter = "notAfter"
	// ImageTagImmutability is a container registry's tag-mutability posture:
	// mutable tags let a pushed image be silently replaced.
	ImageTagImmutability = "imageTagImmutability"
	// ScanOnPush is whether a container registry scans images for vulnerabilities
	// on push.
	ScanOnPush = "scanOnPush"
	// AnonymousPullEnabled is whether a container registry serves image pulls
	// without authentication (a public supply-chain read surface).
	AnonymousPullEnabled = "anonymousPullEnabled"
	// AdminUserEnabled is whether a container registry's shared admin account is
	// enabled (a single long-lived credential for the whole registry).
	AdminUserEnabled = "adminUserEnabled"
	// LocalAuthDisabled is whether shared-key (SAS) authentication is turned off
	// on a messaging namespace, forcing identity-based (Entra) auth. False means
	// long-lived shared keys still grant access alongside RBAC.
	LocalAuthDisabled = "localAuthDisabled"
	// BacksUpSourceIds is a backup vault's recovery-point source resource ids: the
	// BacksUp FK list a vault emits so the graph can join it to every node it holds
	// a copy of (the blast-radius edge "this vault backs up these N stores"). Same id
	// form as each source node's id.
	BacksUpSourceIds = "backsUpSourceIds"
)

var All = []string{
	SourceProvider,
	ResourceType,
	ResourceID,
	AccountID,
	AccountName,
	OrganizationID,
	Region,
	ParentID,
	ParentVisibility,
	State,
	OrgKind,
	AccountKind,
	AltIDs,
	ManagementAccountID,
	ARN,
	ProjectID,
	ProjectName,
	Zone,
	NetworkID,
	CidrBlocks,
	IPv6CidrBlocks,
	IsDefault,
	DNSServers,
	DdosProtected,
	RoutingMode,
	RouteTableID,
	SecurityGroupIDs,
	NatGatewayID,
	AutoAssignPublicIP,
	PrivateGoogleAccess,
	Purpose,
	Kind,
	Routes,
	Rules,
	AssociatedSubnetIDs,
	IsMain,
	SubnetID,
	AttachedInstanceID,
	PrivateIPs,
	PublicIPIDs,
	IPForwardingEnabled,
	AttachedNicIDs,
	AttachedSubnetIDs,
	TargetTags,
	TargetServiceAccounts,
	Scope,
	PrefixListID,
	Cidrs,
	AddressFamily,
	Owner,
	SubnetIDs,
	ConnectivityType,
	ASN,
	HasCloudNat,
	SKU,
	TunnelingEnabled,
	IPConnectEnabled,
	ScaleUnits,
	PolicyID,
	AttachedFirewallIDs,
	DefaultAction,
	FlowLogsEnabled,
	AssociatedNetworkIDs,
	ApplicationSecurityGroupIDs,
	AttachedScopeIds,
	IPAddress,
	Allocation,
	AttachedResourceID,
	Fqdn,
	DomainName,
	Visibility,
	LinkedNetworkIDs,
	ZoneID,
	RecordType,
	TTL,
	Values,
	AliasTargetID,
	LocalNetworkID,
	RemoteNetworkID,
	RemoteAccountID,
	RemoteRegion,
	RemoteCidrBlocks,
	AllowForwardedTraffic,
	AllowGatewayTransit,
	UseRemoteGateways,
	DefaultRouteTableAssociation,
	AddressPrefix,
	HubAttachments,
	EndpointKind,
	NicIDs,
	TargetServiceID,
	PrivateDNSEnabled,
	BackendLbID,
	AcceptanceRequired,
	AllowedConsumers,
	PublicIps,
	GatewayID,
	RemoteSiteID,
	StaticRoutesOnly,
	TunnelCount,
	Bandwidth,
	RemoteAsn,
	AttachedHubOrGatewayID,
	ClientCidrBlocks,
	AuthType,
	SplitTunnel,
	AuthorizationRules,
	Scheme,
	Layer,
	DNSName,
	WafPolicyID,
	CdnEnabled,
	LbID,
	Port,
	Protocol,
	CertificateIDs,
	DefaultActions,
	TargetKind,
	Targets,
	StaticIps,
	Enabled,
	ListenerPorts,
	EndpointTargets,
	RoutingMethod,
	Endpoints,
	MonitorConfig,
	Exposure,
	VpcLinkTargetIDs,
	CustomDomains,
	AuthorizerCount,
	DomainNames,
	OriginHosts,
	OriginIDs,
	ViewerTLSPolicy,
	Mode,
	ManagedRuleSets,
	AttachedResourceIDs,
	PublicNetworkAccess,
	AuthorizedCidrs,
	APIServerPublic,
	APIServerAuthorizedCidrs,
	PrivateNodes,
	PowerState,
	OSType,
	ImageID,
	IdentityIDs,
	IdentityScopes,
	NetworkTags,
	MetadataService,
	InstanceType,
	ImageType,
	LaunchTemplateID,
	BackendPoolIDs,
	Ports,
	NetworkOrigin,
	BucketID,
	PublicURL,
	UrlAuthType,
	Ingress,
	Runtime,
	IdentitySource,
	GroupIds,
	MfaPresent,
	AttachedPolicyIds,
	AccountAssignments,
	PermissionsBoundaryPolicyID,
	PrincipalKind,
	MaxSessionDuration,
	TrustedPrincipals,
	TrustedServices,
	Email,
	Disabled,
	IamBindings,
	AttachmentPointID,
	ResourcePolicyGrants,
	GrantedServices,
	PolicyResourceID,
	RoleID,
	FederationType,
	Issuer,
	ClientIDList,
	ThumbprintList,
	PoolKind,
	PoolID,
	SubjectCondition,
	AttributeMapping,
	AllowUnauthenticatedIdentities,
	RoleTargets,
	HasTokenBasedRoleMapping,
	TenantID,
	PrincipalID,
	RoleDefinitionID,
	ConditionJSON,
	ScheduleStatus,
	RoleName,
	Actions,
	NotActions,
	DataActions,
	NotDataActions,
	ArmIdentityID,
	AppID,
	AppOwnerOrganizationID,
	CredentialCount,
	HasActiveCredential,
	Audiences,
	OwnerApplicationID,
	OwnerArmIdentityID,
	OwnerOrganizationID,
	IdentityProviderKind,
	AttachedIdpIds,
	IdentityPoolTrustRefs,
	ProviderType,
	AppRoles,
	AssigneeID,
	AssigneeType,
	ResourceSpID,
	AppRoleID,
	RoleType,
	Effect,
	Resources,
	NotResources,
	IsNotAction,
	IsNotResource,
	ConditionParsed,
	PolicyRole,
	SourcePolicyID,
	StatementSid,
	AttachedTargets,
	IncludedPermissions,
	DenyRules,
	EncryptionKeyId,
	EnableRbacAuthorization,
	AccessPolicies,
	SourceResourceId,
	Encrypted,
	SharedPublicly,
	TargetBucketID,
	MultiRegion,
	IsOrgTrail,
	LoggingEnabled,
	LogFileValidationEnabled,
	RetentionDays,
	NotAfter,
	ImageTagImmutability,
	ScanOnPush,
	AnonymousPullEnabled,
	AdminUserEnabled,
	LocalAuthDisabled,
	BacksUpSourceIds,
}

var keySet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(All))
	for _, k := range All {
		m[k] = struct{}{}
	}

	return m
}()

func IsCanonical(key string) bool {
	_, ok := keySet[key]

	return ok
}
