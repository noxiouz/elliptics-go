package elliptics

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

type Formatter struct {
	Type    string `json:"type"`
	Pattern string `json:"pattern"`
}

type Rotation struct {
	Move int `json:"move"`
}

type Sink struct {
	Type      string   `json:"type"`
	Path      string   `json:"path"`
	AutoFlush bool     `json:"autoflush"`
	Rotation  Rotation `json:"rotation"`
}

type Frontend struct {
	Formatter Formatter `json:"formatter"`
	Sink      Sink      `json:"sink"`
}

type LoggerCfg struct {
	Frontends []Frontend `json:"frontends"`
	Level     string     `json:"level"`
}

type Monitor struct {
	Port            int32 `json:"port"`
	CallTreeTimeout int32 `json:"call_tree_timeout"`
}

type Cache struct {
	Size uint64 `json:"size"`
}

type Options struct {
	Join                   bool     `json:"join"`
	Flags                  uint64   `json:"flags"`
	Remote                 []string `json:"remote"`
	Address                []string `json:"address"`
	WaitTimeout            uint64   `json:"wait_timeout"`
	CheckTimeout           uint64   `json:"check_timeout"`
	NonBlockingIOThreadNum int      `json:"nonblocking_io_thread_num"`
	IOThreadNum            int      `json:"io_thread_num"`
	NetThreadNum           int      `json:"net_thread_num"`
	Daemon                 bool     `json:"daemon"`
	AuthCookie             string   `json:"auth_cookie"`
	Monitor                Monitor  `json:"monitor"`
}

type BackendCfg struct {
	BackendID        uint32 `json:"backend_id"`
	Type             string `json:"type"`
	Group            uint32 `json:"group"`
	History          string `json:"history"`
	Data             string `json:"data"`
	Sync             int    `json:"sync"`
	BlobFlags        uint64 `json:"blob_flags"`
	BlobSize         string `json:"blob_size"`
	RecordsInBlob    uint64 `json:"records_in_blob"`
	BlobSizeLimit    string `json:"blob_size_limit"`
	DefragPercentage int    `json:"defrag_percentage"`
	PeriodicTimeout  int    `json:"periodic_timeout"`
}

type EllipticsServerConfig struct {
	Logger   LoggerCfg    `json:"logger"`
	Options  Options      `json:"options"`
	Backends []BackendCfg `json:"backends"`
}

func (config *EllipticsServerConfig) Save(file string) error {
	data, err := json.MarshalIndent(config, "", "	")
	if err != nil {
		return fmt.Errorf("Could not marshal elliptics transport config structure: %v", err)
	}

	if err = ioutil.WriteFile(file, data, 0644); err != nil {
		return fmt.Errorf("Could not write file '%s': %v", file, err)
	}

	return nil
}

func (config *EllipticsServerConfig) Load(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, config)
	if err != nil {
		return err
	}

	return nil
}

type DnetIOServ struct {
	cmd    *exec.Cmd
	config *EllipticsServerConfig
	base   string
}

func StartDnetIOServ(groups []uint32) (*DnetIOServ, error) {
	base, err := ioutil.TempDir("", "elliptics-test")
	if err != nil {
		return nil, err
	}

	logfile, err := ioutil.TempFile(base, "elliptics-test-log")
	if err != nil {
		return nil, err
	}
	serverLog := logfile.Name()
	logfile.Close()

	port := rand.Int31n(20000) + 20000

	config := &EllipticsServerConfig{
		Logger: LoggerCfg{
			Frontends: []Frontend{
				Frontend{
					Formatter: Formatter{
						Type:    "string",
						Pattern: "%(timestamp)s %(request_id)s/%(lwp)s/%(pid)s %(severity)s: %(message)s %(...L)s",
					},
					Sink: Sink{
						Type:      "files",
						Path:      serverLog,
						AutoFlush: true,
						Rotation: Rotation{
							Move: 0,
						},
					},
				},
			},
			Level: "notice",
		},
		Options: Options{
			Join:   true,
			Flags:  20,
			Remote: []string{},
			Address: []string{
				fmt.Sprintf("localhost:%d:2-0", port),
			},
			WaitTimeout:            60,
			CheckTimeout:           120,
			NonBlockingIOThreadNum: 4,
			IOThreadNum:            4,
			NetThreadNum:           1,
			Daemon:                 false,
			AuthCookie:             fmt.Sprintf("%016x", rand.Int63()),
			Monitor: Monitor{
				Port: rand.Int31n(20000) + 40000,
			},
		},
		Backends: make([]BackendCfg, 0),
	}

	for i, group := range groups {
		id := i + 1
		backend := BackendCfg{
			BackendID:       uint32(id),
			Type:            "blob",
			Group:           group,
			History:         filepath.Join(base, strconv.Itoa(id), "history"),
			Data:            filepath.Join(base, strconv.Itoa(id), "data/data"),
			Sync:            -1,
			BlobFlags:       0, // bit 4 must not be set to enable blob-size-limit check
			BlobSize:        "20M",
			RecordsInBlob:   1000,
			BlobSizeLimit:   fmt.Sprintf("%dM", 200+rand.Intn(5)*20),
			PeriodicTimeout: 30,
		}

		err = os.MkdirAll(backend.History, 0755)
		if err != nil {
			return nil, fmt.Errorf("Could not create directory '%s': %v", backend.History, err)
		}
		dir := filepath.Dir(backend.Data)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return nil, fmt.Errorf("Could not create directory '%s': %v", dir, err)
		}

		config.Backends = append(config.Backends, backend)
	}

	file := fmt.Sprintf("%s/ioserv.conf", base)
	if err = config.Save(file); err != nil {
		return nil, fmt.Errorf("Could not save config: %v", err)
	}

	ioservCmd := exec.Command("dnet_ioserv", "-c", file)
	if err = ioservCmd.Start(); err != nil {
		return nil, fmt.Errorf("Could not start dnet_ioserv process: %v\n", err)
	}

	// wait 1 second for server to start
	time.Sleep(1 * time.Second)

	// check availability
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		// Cleanup
		ioservCmd.Process.Kill()
		os.RemoveAll(base)
		return nil, err
	}
	defer conn.Close()

	return &DnetIOServ{
		cmd:    ioservCmd,
		config: config,
		base:   base,
	}, nil
}

func (d *DnetIOServ) Close() error {
	defer os.RemoveAll(d.base)
	if err := d.cmd.Process.Kill(); err != nil {
		return err
	}
	d.cmd.Wait()
	return nil
}

func (d *DnetIOServ) Address() []string {
	return d.config.Options.Address
}
