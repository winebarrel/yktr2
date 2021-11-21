package utils

import (
	"html/template"
	"io/fs"
	"io/ioutil"
	"regexp"

	"github.com/gliderlabs/sigil/builtin"
)

var sigilFuncMap = template.FuncMap{
	"append":       builtin.Append,
	"base64decode": builtin.Base64Decode,
	"base64encode": builtin.Base64Encode,
	"capitalize":   builtin.Capitalize,
	"default":      builtin.Default,
	"dir":          builtin.Dir,
	"dirs":         builtin.Dirs,
	"drop":         builtin.Drop,
	"exists":       builtin.Exists,
	"file":         builtin.File,
	"files":        builtin.Files,
	"httpget":      builtin.HttpGet,
	"include":      builtin.Include,
	"indent":       builtin.Indent,
	"jmespath":     builtin.JmesPath,
	"join":         builtin.Join,
	"joinkv":       builtin.JoinKv,
	"json":         builtin.Json,
	"lower":        builtin.Lower,
	"match":        builtin.Match,
	"pointer":      builtin.Pointer,
	"render":       builtin.Render,
	"replace":      builtin.Replace,
	"seq":          builtin.Seq,
	"shell":        builtin.Shell,
	"split":        builtin.Split,
	"splitkv":      builtin.SplitKv,
	"stdin":        builtin.Stdin,
	"substring":    builtin.Substring,
	"text":         builtin.Text,
	"tojson":       builtin.ToJson,
	"toyaml":       builtin.ToYaml,
	"trim":         builtin.Trim,
	"uniq":         builtin.Uniq,
	"upper":        builtin.Upper,
	"var":          builtin.Var,
	"yaml":         builtin.Yaml,
	"emoji":        emoji,
}

func emoji(str string) (interface{}, error) {
	r := regexp.MustCompile(":([[:alnum:]]+):")
	str = r.ReplaceAllString(str, `<img class="emoji" title=":$1:" alt=":$1:" src="https://assets.esa.io/images/emoji/$1.png">`)
	return template.HTML(str), nil
}

func NewTemplate(fs fs.FS, path string) *template.Template {
	file, _ := fs.Open(path)
	defer file.Close()
	content, _ := ioutil.ReadAll(file)
	return template.Must(template.New(path).Funcs(sigilFuncMap).Parse(string(content)))
}
