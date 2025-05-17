package sub

import (
	"context"
	"fmt"
	"github.com/li127103/frp/pkg/config"
	v1 "github.com/li127103/frp/pkg/config/v1"
	clientsdk "github.com/li127103/frp/pkg/sdk/client"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

var adminAPITimeout = 30 * time.Second

func init() {
	commands := []struct {
		name        string
		description string
		handler     func(*v1.ClientCommonConfig) error
	}{
		{"reload", "Hot-Reload frpc configuration", ReloadHandler},
		{"status", "Overview of all proxies status", StatusHandler},
		{"stop", "Stop the running frpc", StopHandler},
	}

	for _, cmdConfig := range commands {
		cmd := NewAdminCommand(cmdConfig.name, cmdConfig.description, cmdConfig.handler)
		cmd.Flags().DurationVar(&adminAPITimeout, "api-timeout", adminAPITimeout, "Timeout for admin API calls")
		rootCmd.AddCommand(cmd)
	}
}

func NewAdminCommand(name, short string, handler func(*v1.ClientCommonConfig) error) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _, _, _, err := config.LoadClientConfig(cfgFile, strictConfigMode)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if cfg.WebServer.Port <= 0 {
				fmt.Println("web server port should be set if you want to use this feature")
				os.Exit(1)
			}

			if err := handler(cfg); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
}

func ReloadHandler(clientCfg *v1.ClientCommonConfig) error {
	client := clientsdk.New(clientCfg.WebServer.Addr, clientCfg.WebServer.Port)
	client.SetAuth(clientCfg.WebServer.User, clientCfg.WebServer.Password)
	ctx, cancel := context.WithTimeout(context.Background(), adminAPITimeout)
	defer cancel()
	if err := client.Reload(ctx, strictConfigMode); err != nil {
		return err
	}
	fmt.Println("reload success")
	return nil
}

func StatusHandler(clientCfg *v1.ClientCommonConfig) error {
	client := clientsdk.New(clientCfg.WebServer.Addr, clientCfg.WebServer.Port)
	client.SetAuth(clientCfg.WebServer.User, clientCfg.WebServer.Password)
	ctx, cancel := context.WithTimeout(context.Background(), adminAPITimeout)
	defer cancel()
	res, err := client.GetAllProxyStatus(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Proxy Status...\n\n")
	for _, typ := range proxyTypes {
		arrs := res[string(typ)]
		if len(arrs) == 0 {
			continue
		}

		fmt.Println(strings.ToUpper(string(typ)))
		tbl := table.New("Name", "Status", "LocalAddr", "Plugin", "RemoteAddr", "Error")
		for _, ps := range arrs {
			tbl.AddRow(ps.Name, ps.Status, ps.LocalAddr, ps.Plugin, ps.RemoteAddr, ps.Err)
		}
		tbl.Print()
		fmt.Println("")
	}
	return nil
}

func StopHandler(clientCfg *v1.ClientCommonConfig) error {
	client := clientsdk.New(clientCfg.WebServer.Addr, clientCfg.WebServer.Port)
	client.SetAuth(clientCfg.WebServer.User, clientCfg.WebServer.Password)
	ctx, cancel := context.WithTimeout(context.Background(), adminAPITimeout)
	defer cancel()
	if err := client.Stop(ctx); err != nil {
		return err
	}
	fmt.Println("stop success")
	return nil
}
