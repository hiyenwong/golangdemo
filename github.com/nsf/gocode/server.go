package main

import (
	"bytes"
	"fmt"
	"go/build"
	"log"
	"net"
	"net/rpc"
	"os"
	"reflect"
	"runtime"
)

func do_server() int {
	g_config.read()
	if g_config.ForceDebugOutput != "" {
		// forcefully enable debugging and redirect logging into the
		// specified file
		*g_debug = true
		f, err := os.Create(g_config.ForceDebugOutput)
		if err != nil {
			panic(err)
		}
		log.SetOutput(f)
	}

	addr := *g_addr
	if *g_sock == "unix" {
		addr = get_socket_filename()
		if file_exists(addr) {
			log.Printf("unix socket: '%s' already exists\n", addr)
			return 1
		}
	}
	g_daemon = new_daemon(*g_sock, addr)
	if *g_sock == "unix" {
		// cleanup unix socket file
		defer os.Remove(addr)
	}

	rpc.Register(new(RPC))

	g_daemon.loop()
	return 0
}

//-------------------------------------------------------------------------
// daemon
//-------------------------------------------------------------------------

type daemon struct {
	listener     net.Listener
	cmd_in       chan int
	autocomplete *auto_complete_context
	pkgcache     package_cache
	declcache    *decl_cache
	context      build.Context
}

func new_daemon(network, address string) *daemon {
	var err error

	d := new(daemon)
	d.listener, err = net.Listen(network, address)
	if err != nil {
		panic(err)
	}

	d.cmd_in = make(chan int, 1)
	d.pkgcache = new_package_cache()
	d.declcache = new_decl_cache(d.context)
	d.autocomplete = new_auto_complete_context(d.pkgcache, d.declcache)
	return d
}

func (this *daemon) drop_cache() {
	this.pkgcache = new_package_cache()
	this.declcache = new_decl_cache(this.context)
	this.autocomplete = new_auto_complete_context(this.pkgcache, this.declcache)
}

const (
	daemon_close = iota
)

func (this *daemon) loop() {
	conn_in := make(chan net.Conn)
	go func() {
		for {
			c, err := this.listener.Accept()
			if err != nil {
				panic(err)
			}
			conn_in <- c
		}
	}()
	for {
		// handle connections or server CMDs (currently one CMD)
		select {
		case c := <-conn_in:
			rpc.ServeConn(c)
			runtime.GC()
		case cmd := <-this.cmd_in:
			switch cmd {
			case daemon_close:
				return
			}
		}
	}
}

func (this *daemon) close() {
	this.cmd_in <- daemon_close
}

var g_daemon *daemon

//-------------------------------------------------------------------------
// server_* functions
//
// Corresponding client_* functions are autogenerated by goremote.
//-------------------------------------------------------------------------

func server_auto_complete(file []byte, filename string, cursor int, context_packed go_build_context) (c []candidate, d int) {
	context := unpack_build_context(&context_packed)
	defer func() {
		if err := recover(); err != nil {
			print_backtrace(err)
			c = []candidate{
				{"PANIC", "PANIC", decl_invalid},
			}

			// drop cache
			g_daemon.drop_cache()
		}
	}()
	// TODO: Probably we don't care about comparing all the fields, checking GOROOT and GOPATH
	// should be enough.
	if !reflect.DeepEqual(g_daemon.context, context) {
		g_daemon.context = context
		g_daemon.drop_cache()
	}
	if *g_debug {
		var buf bytes.Buffer
		log.Printf("Got autocompletion request for '%s'\n", filename)
		log.Printf("Cursor at: %d\n", cursor)
		buf.WriteString("-------------------------------------------------------\n")
		buf.Write(file[:cursor])
		buf.WriteString("#")
		buf.Write(file[cursor:])
		log.Print(buf.String())
		log.Println("-------------------------------------------------------")
	}
	candidates, d := g_daemon.autocomplete.apropos(file, filename, cursor)
	if *g_debug {
		log.Printf("Offset: %d\n", d)
		log.Printf("Number of candidates found: %d\n", len(candidates))
		log.Printf("Candidates are:\n")
		for _, c := range candidates {
			abbr := fmt.Sprintf("%s %s %s", c.Class, c.Name, c.Type)
			if c.Class == decl_func {
				abbr = fmt.Sprintf("%s %s%s", c.Class, c.Name, c.Type[len("func"):])
			}
			log.Printf("  %s\n", abbr)
		}
		log.Println("=======================================================")
	}
	return candidates, d
}

func server_cursor_type_pkg(file []byte, filename string, cursor int) (typ, pkg string) {
	defer func() {
		if err := recover(); err != nil {
			print_backtrace(err)

			// drop cache
			g_daemon.drop_cache()
		}
	}()
	return g_daemon.autocomplete.cursor_type_pkg(file, filename, cursor)
}

func server_close(notused int) int {
	g_daemon.close()
	return 0
}

func server_status(notused int) string {
	return g_daemon.autocomplete.status()
}

func server_drop_cache(notused int) int {
	// drop cache
	g_daemon.drop_cache()
	return 0
}

func server_set(key, value string) string {
	if key == "\x00" {
		return g_config.list()
	} else if value == "\x00" {
		return g_config.list_option(key)
	}
	return g_config.set_option(key, value)
}
