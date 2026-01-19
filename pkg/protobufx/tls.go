package protobufx

import (
	"crypto/tls"
	"crypto/x509"
	"os"
)

func (x *TLSOptions) AsConfig() *tls.Config {
	if x == nil {
		return nil
	}
	cert, err := tls.LoadX509KeyPair(x.GetCertFile().GetValue(), x.GetKeyFile().GetValue())
	if err != nil {
		panic(err)
	}
	// 创建基础 TLS 配置
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	// 处理 CA 证书
	if x.GetCaFile() != nil {
		caCert, err := os.ReadFile(x.GetCaFile().GetValue())
		if err != nil {
			panic(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	// 设置服务器名称
	if x.GetServerName() != nil {
		tlsConfig.ServerName = x.GetServerName().GetValue()
	}
	return tlsConfig
}
