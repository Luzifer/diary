package main

//go:generate go-bindata -o assets.go assets/

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/Luzifer/go-openssl"
	"github.com/bgentry/speakeasy"
	"github.com/spf13/cobra"
)

var (
	version = "dev"

	set    *settings
	params = struct {
		SettingsFile string
	}{}
	tmpFile string
	tmpPass string

	rootCmd *cobra.Command
)

func init() {
	rootCmd = &cobra.Command{
		Use: "diary",
	}

	rootCmd.AddCommand([]*cobra.Command{{
		Use:               "add",
		Short:             "Creates a new daily entry in the diary",
		PersistentPreRunE: loadSettings,
		PreRun:            preDecrypt,
		Run:               actionAdd,
		PostRun:           postEncrypt,
	}, {
		Use:               "edit",
		Short:             "Opens the editor with the diary file",
		PersistentPreRunE: loadSettings,
		PreRun:            preDecrypt,
		Run:               actionEdit,
		PostRun:           postEncrypt,
	}, {
		Use:   "init",
		Short: "Copies settings.yml and template.md examples into storing directory",
		Run:   actionInit,
	}}...)

	currentUser, err := user.Current()
	if err != nil {
		log.Println("Could not get current user")
		os.Exit(1)
	}
	defaultSettings := path.Join(currentUser.HomeDir, ".config", "diary", "settings.yml")
	rootCmd.Flags().StringVar(&params.SettingsFile, "settings", defaultSettings, "Where to store the settings")
}

func main() {
	rootCmd.Execute()
}

func preDecrypt(cmd *cobra.Command, args []string) {
	p := path.Dir(params.SettingsFile)
	diaryStore, err := ioutil.ReadFile(path.Join(p, "diary.md"))
	if err != nil {
		log.Printf("Unable to load stored diary")
		os.Exit(1)
	}
	tmp, err := ioutil.TempFile("", "diary")
	if err != nil {
		log.Printf("Unable to open a temp file")
		os.Exit(1)
	}
	tmpFile = tmp.Name()

	if set.Encrypt {
		pwd, err := speakeasy.Ask("Password: ")
		if err != nil {
			log.Printf("Unable to read password.")
			os.Exit(1)
		}
		tmpPass = pwd

		if len(diaryStore) > 0 {
			o := openssl.New()
			diaryStore, err = o.DecryptString(tmpPass, string(diaryStore))
			if err != nil {
				log.Printf("Unable to decrypt diary")
				os.Exit(1)
			}
		}
	}

	fmt.Fprintf(tmp, string(diaryStore))
	tmp.Close()
}

func postEncrypt(cmd *cobra.Command, args []string) {
	p := path.Dir(params.SettingsFile)
	diaryStore, err := ioutil.ReadFile(tmpFile)
	if err != nil {
		log.Printf("Unable to load stored diary")
		os.Remove(tmpFile)
		os.Exit(1)
	}

	if set.Encrypt {
		o := openssl.New()
		diaryStore, err = o.EncryptString(tmpPass, string(diaryStore))
	}

	err = ioutil.WriteFile(path.Join(p, "diary.md"), diaryStore, 0600)
	if err != nil {
		log.Printf("Unable to store the diary")
		os.Remove(tmpFile)
		os.Exit(1)
	}

	os.Remove(tmpFile)
}

func actionAdd(cmd *cobra.Command, args []string) {
	dateString := time.Now().Format(set.DateFormat)
	diary, err := ioutil.ReadFile(tmpFile)
	if err != nil {
		log.Printf("Unable to read diary content")
		os.Remove(tmpFile)
		os.Exit(1)
	}

	if strings.Contains(string(diary), dateString) {
		log.Printf("The diary already contains the current date.")
		os.Remove(tmpFile)
		os.Exit(1)
	}

	p := path.Dir(params.SettingsFile)
	tplText, err := ioutil.ReadFile(path.Join(p, "template.md"))
	if err != nil {
		log.Printf("Unable to read template.md")
		os.Remove(tmpFile)
		os.Exit(1)
	}

	tpl, err := template.New("template").Parse(fmt.Sprintf("%s\n\n", string(tplText)))
	if err != nil {
		log.Printf("Unable to parse template.md")
		os.Remove(tmpFile)
		os.Exit(1)
	}

	buf := bytes.NewBuffer([]byte{})
	err = tpl.Execute(buf, struct{ Date string }{Date: dateString})
	if err != nil {
		log.Printf("Unable to execute template.md")
		os.Remove(tmpFile)
		os.Exit(1)
	}

	buf.Write(diary)
	err = ioutil.WriteFile(tmpFile, buf.Bytes(), 0600)
	if err != nil {
		log.Printf("Unable to add template")
		os.Remove(tmpFile)
		os.Exit(1)
	}
}

func actionEdit(cmd *cobra.Command, args []string) {
	tpl, err := template.New("EditCmd").Parse(set.EditorCmd)
	if err != nil {
		log.Printf("Unable to read EditorCmd")
		os.Remove(tmpFile)
		os.Exit(1)
	}

	buf := bytes.NewBuffer([]byte{})
	err = tpl.Execute(buf, struct{ DiaryFile string }{DiaryFile: tmpFile})
	if err != nil {
		log.Printf("Unable to parse EditorCmd")
		os.Remove(tmpFile)
		os.Exit(1)
	}

	cStr := strings.Split(buf.String(), " ")
	c := exec.Command(cStr[0], cStr[1:]...)
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout

	err = c.Run()
	if err != nil {
		log.Printf("Editor signaled there was an error, not saving")
		os.Remove(tmpFile)
		os.Exit(1)
	}
}

func actionInit(cmd *cobra.Command, args []string) {
	if _, err := os.Stat(params.SettingsFile); err == nil {
		log.Printf("Settings file already exists, will not overwrite")
		return
	}

	p := path.Dir(params.SettingsFile)
	os.MkdirAll(p, 0700)

	setData, _ := Asset("assets/settings.yml")
	ioutil.WriteFile(params.SettingsFile, setData, 0600)

	temData, _ := Asset("assets/template.md")
	ioutil.WriteFile(path.Join(p, "template.md"), temData, 0600)

	ioutil.WriteFile(path.Join(p, "diary.md"), []byte{}, 0600)

	log.Printf("Settings file and Template copied to %s", p)
}
