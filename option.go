package xlistener

import (
	"fmt"
	"os"
	"time"
)

//go:generate gotemplate -outfmt gen_%v "bitbucket.org/funplus/sandwich/base/container/templates/smap" "IntFdInfo(int,*fdInfo,func(k int) uint64     { return uint64(k) })"
//go:generate optiongen --xconf=true --usage_tag_name=usage
func ConfOptionDeclareWithDefault() interface{} {
	return map[string]interface{}{
		// annotation@BacklogAccept(comment="like Linux TCP backlog")
		"BacklogAccept": 1024,
		// annotation@TimeoutCanRead(comment="conn will be closed if no data arrived after TimeoutCanRead since conn established")
		"TimeoutCanRead": time.Duration(time.Second * time.Duration(30)),
		// annotation@EnableHandshake(comment="enable handler shake between server and client")
		"EnableHandshake": false,
		// annotation@HandshakeTimeout(comment="if can not finish Handshake withen HandshakeTimeout, conn will be closed")
		"HandshakeTimeout": time.Duration(time.Second * time.Duration(10)),
		// annotation@Debugf(comment="debug log func")
		"Debugf": func(format string, v ...interface{}) { fmt.Fprintf(os.Stdout, format, v...) },
		// annotation@Warningf(comment="warn log func")
		"Warningf": func(format string, v ...interface{}) { fmt.Fprintf(os.Stderr, format, v...) },
	}
}
