package validation

import (
	"fmt"
	v1 "github.com/li127103/frp/pkg/config/v1"
	"github.com/li127103/frp/pkg/featuregate"
	"github.com/samber/lo"
	"os"
	"path/filepath"
	"slices"
)

func ValidateClientCommonConfig(c *v1.ClientCommonConfig) (Warning, error) {
	var (
		warnings Warning
		errs     error
	)
	// validate feature gates
	if c.VirtualNet.Address != "" {
		if !featuregate.Enabled(featuregate.VirtualNet) {
			return warnings, fmt.Errorf("VirtualNet feature is not enabled; enable it by setting the appropriate feature gate flag")
		}
	}

	if !slices.Contains(SupportedAuthMethods, c.Auth.Method) {
		errs = AppendError(errs, fmt.Errorf("invalid auth method, optional values are %v", SupportedAuthMethods))
	}
	if !lo.Every(SupportedAuthAdditionalScopes, c.Auth.AdditionalScopes) {
		errs = AppendError(errs, fmt.Errorf("invalid auth additional scopes, optional values are %v", SupportedAuthAdditionalScopes))
	}

	if err := validateLogConfig(&c.Log); err != nil {
		errs = AppendError(errs, err)
	}

	if err := validateWebServerConfig(&c.WebServer); err != nil {
		errs = AppendError(errs, err)
	}

	if c.Transport.HeartbeatTimeout > 0 && c.Transport.HeartbeatInterval > 0 {
		if c.Transport.HeartbeatTimeout < c.Transport.HeartbeatInterval {
			errs = AppendError(errs, fmt.Errorf("invalid transport.heartbeatTimeout, heartbeat timeout should not less than heartbeat interval"))
		}
	}

	if !lo.FromPtr(c.Transport.TLS.Enable) {
		checkTLSConfig := func(name string, value string) Warning {
			if value != "" {
				return fmt.Errorf("%s is invalid when transport.tls.enable is false", name)
			}
			return nil
		}

		warnings = AppendError(warnings, checkTLSConfig("transport.tls.certFile", c.Transport.TLS.CertFile))
		warnings = AppendError(warnings, checkTLSConfig("transport.tls.keyFile", c.Transport.TLS.KeyFile))
		warnings = AppendError(warnings, checkTLSConfig("transport.tls.trustedCaFile", c.Transport.TLS.TrustedCaFile))
	}

	if !slices.Contains(SupportedTransportProtocols, c.Transport.Protocol) {
		errs = AppendError(errs, fmt.Errorf("invalid transport.protocol, optional values are %v", SupportedTransportProtocols))
	}

	for _, f := range c.IncludeConfigFiles {
		absDir, err := filepath.Abs(filepath.Dir(f))
		if err != nil {
			errs = AppendError(errs, fmt.Errorf("include: parse directory of %s failed: %v", f, err))
			continue
		}
		if _, err := os.Stat(absDir); os.IsNotExist(err) {
			errs = AppendError(errs, fmt.Errorf("include: directory of %s not exist", f))
		}
	}
	return warnings, errs
}

func ValidateAllClientConfig(c *v1.ClientCommonConfig, proxyCfgs []v1.ProxyConfigurer, visitorCfgs []v1.VisitorConfigurer) (Warning, error) {
	var warnings Warning
	if c != nil {
		warning, err := ValidateClientCommonConfig(c)
		warnings = AppendError(warnings, warning)
		if err != nil {
			return warnings, err
		}
	}

	for _, c := range proxyCfgs {
		if err := ValidateProxyConfigurerForClient(c); err != nil {
			return warnings, fmt.Errorf("proxy %s: %v", c.GetBaseConfig().Name, err)
		}
	}

	for _, c := range visitorCfgs {
		if err := ValidateVisitorConfigurer(c); err != nil {
			return warnings, fmt.Errorf("visitor %s: %v", c.GetBaseConfig().Name, err)
		}
	}
	return warnings, nil
}
