package sub

import (
	"fmt"
	"github.com/li127103/frp/pkg/config"
	v1 "github.com/li127103/frp/pkg/config/v1"
	"github.com/li127103/frp/pkg/config/v1/validation"
	"github.com/spf13/cobra"
	"os"
	"slices"
)

var proxyTypes = []v1.ProxyType{
	v1.ProxyTypeTCP,
	v1.ProxyTypeUDP,
	v1.ProxyTypeTCPMUX,
	v1.ProxyTypeHTTP,
	v1.ProxyTypeHTTPS,
	v1.ProxyTypeSTCP,
	v1.ProxyTypeSUDP,
	v1.ProxyTypeXTCP,
}

var visitorTypes = []v1.VisitorType{
	v1.VisitorTypeSTCP,
	v1.VisitorTypeSUDP,
	v1.VisitorTypeXTCP,
}

func init() {
	for _, typ := range proxyTypes {
		c := v1.NewProxyConfigurerByType(typ)
		if c == nil {
			panic("proxy type: " + typ + " not support")
		}
		clientCfg := v1.ClientCommonConfig{}
		cmd := NewProxyCommand(string(typ), c, &clientCfg)
		config.RegisterClientCommonConfigFlags(cmd, &clientCfg)
		config.RegisterProxyFlags(cmd, c)

		// add sub command for visitor
		if slices.Contains(visitorTypes, v1.VisitorType(typ)) {
			vc := v1.NewVisitorConfigurerByType(v1.VisitorType(typ))
			if vc == nil {
				panic("visitor type: " + typ + " not support")
			}
			visitorCmd := NewVisitorCommand(string(typ), vc, &clientCfg)
			config.RegisterVisitorFlags(visitorCmd, vc)
			cmd.AddCommand(visitorCmd)
		}
		rootCmd.AddCommand(cmd)
	}
}

func NewProxyCommand(name string, c v1.ProxyConfigurer, clientCfg *v1.ClientCommonConfig) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: fmt.Sprintf("Run frpc with a single %s proxy", name),
		Run: func(cmd *cobra.Command, args []string) {
			clientCfg.Complete()
			if _, err := validation.ValidateClientCommonConfig(clientCfg); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			c.Complete(clientCfg.User)
			c.GetBaseConfig().Type = name
			if err := validation.ValidateProxyConfigurerForClient(c); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err := startService(clientCfg, []v1.ProxyConfigurer{c}, nil, "")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
}

func NewVisitorCommand(name string, c v1.VisitorConfigurer, clientCfg *v1.ClientCommonConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "visitor",
		Short: fmt.Sprintf("Run frpc with a single %s visitor", name),
		Run: func(cmd *cobra.Command, args []string) {
			clientCfg.Complete()
			if _, err := validation.ValidateClientCommonConfig(clientCfg); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			c.Complete(clientCfg)
			c.GetBaseConfig().Type = name
			if err := validation.ValidateVisitorConfigurer(c); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err := startService(clientCfg, nil, []v1.VisitorConfigurer{c}, "")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
}
