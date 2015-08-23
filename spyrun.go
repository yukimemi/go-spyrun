package spyrun

import ( // {{{
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/naoina/toml"
) // }}}

const ( // {{{
	// SpyRunFile convert to target file.
	SpyRunFile = "\\$SPYRUN_FILE"
) // }}}

/**
 * Toml config.
 */
type tomlConfig struct { // {{{
	Spyconf struct {
		Sleep string `toml:"sleep"`
	}
	SpyTables map[string]spyTable `toml:"spys"`
} // }}}

type spyTable struct { // {{{
	File    string `toml:"file"`
	Command string `toml:"command"`
} // }}}

/**
 * Spyrun config.
 */
type spyMap map[string][]*spyst

type spyst struct { // {{{
	filePath   string
	command    string
	modifyTime time.Time
	mu         *sync.Mutex
} // }}}

type spyrun struct { // {{{
	conf tomlConfig
	spym spyMap
} // }}}

// New Create and return *spyrun.
func New() *spyrun { // {{{
	s := new(spyrun)
	s.spym = make(spyMap)
	return s
} // }}}

// Run run spyrun.
func Run(tomlpath string) error { // {{{
	return New().run(tomlpath)
} // }}}

func (s *spyrun) run(tomlpath string) error { // {{{
	var err error

	err = s.loadToml(tomlpath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse toml ! %s", err.Error())
		os.Exit(1)
	}

	err = s.createSpyMapFromSpyTables()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get spys ! %s", err.Error())
		os.Exit(1)
	}

	ch := make(chan *spyst)
	go s.spyFiles(ch)

	for {
		spyst := <-ch
		log.Printf("[%s] is modified !\n", spyst.filePath)
		go s.executeCommand(spyst)
	}

} // }}}

func (s *spyrun) convertSpyVar(file, command string) (string, error) { // {{{
	var err error

	re := regexp.MustCompile(SpyRunFile)

	if matched := re.MatchString(command); matched {
		command = re.ReplaceAllString(command, file)
	}

	return command, err
} // }}}

func (s *spyrun) createSpyMapFromSpyTables() error { // {{{
	var err error

	for k, v := range s.conf.SpyTables {
		s.spym[k] = make([]*spyst, 0)
		files, err := filepath.Glob(v.File)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to search glob pattern. %s", v.File)
			os.Exit(1)
		}
		for _, file := range files {
			spyst := new(spyst)
			spyst.filePath = file
			fi, err := os.Stat(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get FileInfo. %s [%s]", file, err.Error())
				os.Exit(1)
			}
			spyst.modifyTime = fi.ModTime()
			spyst.command, err = s.convertSpyVar(file, v.Command)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to convert spy variable. %s", v.Command)
				os.Exit(1)
			}
			spyst.mu = new(sync.Mutex)
			log.Printf("%s: {file: [%s], command: [%s]}\n", k, spyst.filePath, spyst.command)
			s.spym[k] = append(s.spym[k], spyst)
		}
	}
	return err
} // }}}

func (s *spyrun) loadToml(tomlpath string) error { // {{{
	var err error

	if _, err = os.Stat(tomlpath); err != nil {
		fmt.Fprintf(os.Stderr, "%s is not found !", tomlpath)
		os.Exit(1)
	}

	f, err := os.Open(tomlpath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open %s", tomlpath)
		os.Exit(1)
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load %s", tomlpath)
		os.Exit(1)
	}

	err = toml.Unmarshal(buf, &s.conf)

	return err
} // }}}

func (s *spyrun) spyFiles(ch chan *spyst) { // {{{
	var err error
	var sleep time.Duration
	if s.conf.Spyconf.Sleep != "" {
		sleep, err = time.ParseDuration(s.conf.Spyconf.Sleep)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse sleep duration. %v", err)
			os.Exit(1)
		}
	} else {
		sleep = time.Duration(100) * time.Millisecond
	}
	log.Println("sleep:", sleep)
	for {
		for _, spysts := range s.spym {
			for _, spyst := range spysts {
				fi, err := os.Stat(spyst.filePath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to get FileInfo. %s, [%s]", spyst.filePath, err.Error())
					ch <- spyst
					// todo remove file from s.spym
				} else if fi.ModTime() != spyst.modifyTime {
					spyst.modifyTime = fi.ModTime()
					ch <- spyst
				}
			}
		}
		time.Sleep(sleep)
	}
} // }}}

func (s *spyrun) executeCommand(spy *spyst) error { // {{{
	var err error
	spy.mu.Lock()
	defer spy.mu.Unlock()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", spy.command)
	} else {
		cmd = exec.Command("sh", "-c", spy.command)
	}
	log.Printf("Execute command. [%s]", spy.command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

	return err
} // }}}
