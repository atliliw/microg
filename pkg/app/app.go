// Package app 提供基于 Cobra 的 CLI 应用框架
// 用于处理命令行参数、配置文件加载、版本信息等
package app

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cliflag "microg/pkg/common/cli/flag"
	"microg/pkg/common/cli/globalflag"
	"microg/pkg/common/term"
	"microg/pkg/common/version"
	"microg/pkg/common/version/verflag"
	"microg/pkg/errors"
	"microg/pkg/log"
)

var (
	progressMessage = color.GreenString("==>")
	//nolint: deadcode,unused,varcheck
	usageTemplate = fmt.Sprintf(`%s{{if .Runnable}}
  %s{{end}}{{if .HasAvailableSubCommands}}
  %s{{end}}{{if gt (len .Aliases) 0}}

%s
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

%s
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

%s{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  %s {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

%s
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

%s
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

%s{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "%s --help" for more information about a command.{{end}}
`,
		color.CyanString("Usage:"),
		color.GreenString("{{.UseLine}}"),
		color.GreenString("{{.CommandPath}} [command]"),
		color.CyanString("Aliases:"),
		color.CyanString("Examples:"),
		color.CyanString("Available Commands:"),
		color.GreenString("{{rpad .Name .NamePadding }}"),
		color.CyanString("Flags:"),
		color.CyanString("Global Flags:"),
		color.CyanString("Additional help topics:"),
		color.GreenString("{{.CommandPath}} [command]"),
	)
)

// App 是 CLI 应用的核心结构体
// 封装了 Cobra 命令，提供配置加载、参数验证等功能
type App struct {
	basename    string               // 二进制文件名，用于查找配置文件
	name        string               // 应用名称
	description string               // 应用描述
	options     CliOptions           // 配置对象，实现 Flags() 和 Validate() 方法
	runFunc     RunFunc              // 启动回调函数，配置加载完成后执行
	silence     bool                 // 静默模式，不打印启动信息
	noVersion   bool                 // 是否禁用版本标志
	noConfig    bool                 // 是否禁用配置文件
	commands    []*Command           // 子命令列表
	args        cobra.PositionalArgs // 非标志参数验证函数
	cmd         *cobra.Command       // Cobra 命令实例
}

// Option 是应用配置函数，用于修改 App 的字段
type Option func(*App)

// WithOptions 注入配置对象
// 配置对象需要实现 CliOptions 接口（Flags() 和 Validate() 方法）
func WithOptions(opt CliOptions) Option {
	return func(a *App) {
		a.options = opt
	}
}

// RunFunc 定义启动回调函数类型
type RunFunc func(basename string) error

// WithRunFunc 设置启动回调函数
// 该函数在配置加载、验证完成后执行
func WithRunFunc(run RunFunc) Option {
	return func(a *App) {
		a.runFunc = run
	}
}

// WithDescription 设置应用描述
func WithDescription(desc string) Option {
	return func(a *App) {
		a.description = desc
	}
}

// WithSilence 设置静默模式
// 静默模式下不打印启动信息、配置信息、版本信息
func WithSilence() Option {
	return func(a *App) {
		a.silence = true
	}
}

// WithNoVersion 禁用版本标志
func WithNoVersion() Option {
	return func(a *App) {
		a.noVersion = true
	}
}

// WithNoConfig 禁用配置文件读取
func WithNoConfig() Option {
	return func(a *App) {
		a.noConfig = true
	}
}

// WithValidArgs 设置非标志参数验证函数
func WithValidArgs(args cobra.PositionalArgs) Option {
	return func(a *App) {
		a.args = args
	}
}

// WithDefaultValidArgs 设置默认的非标志参数验证
// 禁止传入任何非标志参数
func WithDefaultValidArgs() Option {
	return func(a *App) {
		a.args = func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		}
	}
}

// NewApp 创建 CLI 应用实例
// name: 应用名称
// basename: 二进制文件名，用于查找配置文件（如 "microg" 会查找 microg.yaml）
// opts: 可选配置项
func NewApp(name string, basename string, opts ...Option) *App {
	a := &App{
		name:     name,
		basename: basename,
	}

	// 应用所有配置项
	for _, o := range opts {
		o(a)
	}

	// 构建 Cobra 命令
	a.buildCommand()

	return a
}

// buildCommand 构建 Cobra 命令
// 将 App 的配置转换为 Cobra 命令结构
func (a *App) buildCommand() {
	cmd := cobra.Command{
		Use:   FormatBaseName(a.basename),
		Short: a.name,
		Long:  a.description,
		// 出错时不打印用法
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          a.args,
	}

	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.Flags().SortFlags = true
	cliflag.InitFlags(cmd.Flags())

	// 注册子命令
	if len(a.commands) > 0 {
		for _, command := range a.commands {
			cmd.AddCommand(command.cobraCommand())
		}
		cmd.SetHelpCommand(helpCommand(a.name))
	}

	// 设置主命令的执行函数
	if a.runFunc != nil {
		cmd.RunE = a.runCommand
	}

	// 从配置对象获取命令行参数定义
	var namedFlagSets cliflag.NamedFlagSets
	if a.options != nil {
		namedFlagSets = a.options.Flags()
		fs := cmd.Flags()
		for _, f := range namedFlagSets.FlagSets {
			fs.AddFlagSet(f)
		}

		// 设置帮助和用法输出格式
		usageFmt := "Usage:\n  %s\n"
		cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
		cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
			cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
		})
		cmd.SetUsageFunc(func(cmd *cobra.Command) error {
			fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine())
			cliflag.PrintSections(cmd.OutOrStderr(), namedFlagSets, cols)
			return nil
		})
	}

	// 添加版本标志
	if !a.noVersion {
		verflag.AddFlags(namedFlagSets.FlagSet("global"))
	}

	// 添加配置文件标志
	if !a.noConfig {
		addConfigFlag(a.basename, namedFlagSets.FlagSet("global"))
	}

	// 添加全局标志
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name())

	a.cmd = &cmd
}

// Run 启动 CLI 应用
// 执行 Cobra 命令，开始整个流程
func (a *App) Run() {
	if err := a.cmd.Execute(); err != nil {
		fmt.Printf("%v %v\n", color.RedString("Error:"), err)
		os.Exit(1)
	}
}

// Command 返回内部的 Cobra 命令实例
func (a *App) Command() *cobra.Command {
	return a.cmd
}

// runCommand 是 Cobra 命令的实际执行函数
// 完整流程：打印信息 → 加载配置 → 验证配置 → 执行启动函数
func (a *App) runCommand(cmd *cobra.Command, args []string) error {
	// 打印当前工作目录
	printWorkingDir()
	// 打印命令行参数
	cliflag.PrintFlags(cmd.Flags())

	// 处理版本标志（--version）
	if !a.noVersion {
		verflag.PrintAndExitIfRequested()
	}

	// 加载配置文件
	if !a.noConfig {
		// 绑定命令行参数到 viper
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}

		// 将配置文件内容解析到配置对象
		if err := viper.Unmarshal(a.options); err != nil {
			return err
		}
	}

	// 打印启动信息
	if !a.silence {
		log.Infof("%v Starting %s ...", progressMessage, a.name)
		if !a.noVersion {
			log.Infof("%v Version: `%s`", progressMessage, version.Get().ToJSON())
		}
		if !a.noConfig {
			log.Infof("%v Config file used: `%s`", progressMessage, viper.ConfigFileUsed())
		}
	}

	// 应用配置规则：补全默认值 → 验证配置 → 打印配置
	if a.options != nil {
		if err := a.applyOptionRules(); err != nil {
			return err
		}
	}

	// 执行启动回调函数
	if a.runFunc != nil {
		return a.runFunc(a.basename)
	}

	return nil
}

// applyOptionRules 应用配置规则
// 1. Complete(): 补全默认值
// 2. Validate(): 验证配置合法性
// 3. String(): 打印配置信息
func (a *App) applyOptionRules() error {
	// 补全默认值（如果配置对象实现了 CompleteableOptions 接口）
	if completeableOptions, ok := a.options.(CompleteableOptions); ok {
		if err := completeableOptions.Complete(); err != nil {
			return err
		}
	}

	// 验证配置合法性
	// 如果返回错误列表不为空，则启动失败
	if errs := a.options.Validate(); len(errs) != 0 {
		return errors.NewAggregate(errs)
	}

	// 打印配置信息（如果配置对象实现了 PrintableOptions 接口）
	if printableOptions, ok := a.options.(PrintableOptions); ok && !a.silence {
		log.Infof("%v Config: `%s`", progressMessage, printableOptions.String())
	}

	return nil
}

// printWorkingDir 打印当前工作目录
func printWorkingDir() {
	wd, _ := os.Getwd()
	log.Infof("%v WorkingDir: %s", progressMessage, wd)
}
