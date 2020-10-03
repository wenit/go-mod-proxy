package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/goproxy/goproxy"
	"github.com/spf13/cobra"
	"github.com/wenit/go-mod-proxy/internal/cacher"
	"github.com/wenit/go-mod-proxy/internal/version"
	"github.com/wenit/go-mod-proxy/pkg/common"
)

// 参数
var (
	help       bool   // 打印帮助信息
	ver        bool   // 打印版本信息
	debug      bool   // 开启调试模式
	repository string // 本地仓库目录
	proxyPort  int    // 代理端口
	proxyHost  string // 代理主机名
)

func main() {
	Execute()
}

func startProxy() {
	proxy := goproxy.New()
	proxy.Cacher = &cacher.Disk{
		Root: repository,
	}

	bindAddress := fmt.Sprintf("%s:%d", proxyHost, proxyPort)

	log.Printf("开始启动服务，监听地址：%s", bindAddress)
	err := http.ListenAndServe(bindAddress, proxy)

	if err != nil {
		log.Fatalf("启动代理[%s]服务失败:%v", repository, err)
	}
}

// 根命令
var rootCmd = &cobra.Command{
	DisableFlagsInUseLine: true,
	Use:                   os.Args[0],
	Short:                 "使用说明",
	Long:                  `使用说明：`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if ver {
			fmt.Println(version.GetVersion())
			return
		}

		initRepo()
		startProxy()
	},
}

func initRepo() error {
	err := common.MkDirs(repository)

	if err != nil {
		log.Fatalf("初始化本地仓库目录[%s]失败:%v", repository, err)
	}
	absRepo, _ := filepath.Abs(repository)
	log.Printf("本地仓库目录：%s", absRepo)

	return nil
}

// Execute 执行程序
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	rootCmd.Flags().BoolVar(&help, "help", false, "帮助信息")
	rootCmd.Flags().BoolVarP(&ver, "version", "v", false, "版本信息")
	rootCmd.Flags().StringVarP(&repository, "repository", "r", "./data", "本地仓库目录")
	rootCmd.Flags().StringVarP(&proxyHost, "host", "h", "", "绑定的host")
	rootCmd.Flags().IntVarP(&proxyPort, "port", "p", 8081, "代理端口")

	// 帮助文档
	rootCmd.SetHelpCommand(helpCmd)
}

var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "更多帮助文档",
	Long:  `Help provides help for any command in the application.`,

	Run: func(c *cobra.Command, args []string) {
		cmd, _, e := c.Root().Find(args)
		if cmd == nil || e != nil {
			c.Printf("Unknown help topic %#q\n", args)
			c.Root().Usage()
		} else {
			cmd.InitDefaultHelpFlag() // make possible 'help' flag to be shown
			cmd.Help()
		}
	},
}
