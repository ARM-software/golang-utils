package casing

// InitialismRules defines replacement rules for the common Go initialisms.
//
// This variable can be passed directly to [NewReplacer] to opt into built-in
// Go-style initialism normalisation, for example:
//
//	r, err := NewReplacer(InitialismRules...)
//
// The list is based on the standard Go initialism guidance:
// https://go.dev/wiki/CodeReviewComments#initialisms
var InitialismRules = []Rule{
	{Token: "Acl", Replacement: "ACL"},
	{Token: "Api", Replacement: "API"},
	{Token: "Ascii", Replacement: "ASCII"},
	{Token: "Cpu", Replacement: "CPU"},
	{Token: "Css", Replacement: "CSS"},
	{Token: "Dns", Replacement: "DNS"},
	{Token: "Eof", Replacement: "EOF"},
	{Token: "Guid", Replacement: "GUID"},
	{Token: "Html", Replacement: "HTML"},
	{Token: "Http", Replacement: "HTTP"},
	{Token: "Https", Replacement: "HTTPS"},
	{Token: "Id", Replacement: "ID"},
	{Token: "Ip", Replacement: "IP"},
	{Token: "Json", Replacement: "JSON"},
	{Token: "Lhs", Replacement: "LHS"},
	{Token: "Qps", Replacement: "QPS"},
	{Token: "Ram", Replacement: "RAM"},
	{Token: "Rhs", Replacement: "RHS"},
	{Token: "Rpc", Replacement: "RPC"},
	{Token: "Sla", Replacement: "SLA"},
	{Token: "Smtp", Replacement: "SMTP"},
	{Token: "Sql", Replacement: "SQL"},
	{Token: "Ssh", Replacement: "SSH"},
	{Token: "Tcp", Replacement: "TCP"},
	{Token: "Tls", Replacement: "TLS"},
	{Token: "Ttl", Replacement: "TTL"},
	{Token: "Udp", Replacement: "UDP"},
	{Token: "Ui", Replacement: "UI"},
	{Token: "Uid", Replacement: "UID"},
	{Token: "Uuid", Replacement: "UUID"},
	{Token: "Uri", Replacement: "URI"},
	{Token: "Url", Replacement: "URL"},
	{Token: "Utf8", Replacement: "UTF8"},
	{Token: "Vm", Replacement: "VM"},
	{Token: "Xml", Replacement: "XML"},
	{Token: "Xmpp", Replacement: "XMPP"},
	{Token: "Xsrf", Replacement: "XSRF"},
	{Token: "Xss", Replacement: "XSS"},
}

// InitialismReplacer is a ready-to-use replacer built from [InitialismRules].
//
// This variable provides built-in optional Go initialism support without
// requiring callers to construct a replacer explicitly.
var InitialismReplacer = mustNewReplacer(InitialismRules...)

func mustNewReplacer(rules ...Rule) *Replacer {
	replacer, err := NewReplacer(rules...)
	if err != nil {
		panic(err)
	}
	return replacer
}
