package aes

import (
	"os"

	"github.com/leon-yc/ggs/internal/pkg/goplugin"
	"github.com/leon-yc/ggs/internal/security"
	"github.com/leon-yc/ggs/pkg/qlog"
	security2 "github.com/go-chassis/foundation/security"
)

const cipherPlugin = "cipher_plugin.so"

//Cipher interface declares Init(), Encrypyt(), Decrypyt() methods
type Cipher interface {
	Init()
	Encrypt(src string) (string, error)
	Decrypt(src string) (string, error)
}

// HWAESCipher is a cipher used in huawei
type HWAESCipher struct {
	gcryptoEngine Cipher
}

func init() {
	if v, exist := os.LookupEnv("CIPHER_ROOT"); exist {
		err := os.Setenv("PAAS_CRYPTO_PATH", v)
		if err != nil {
			qlog.Warn("can not set env for cipher: " + err.Error())
		}
	}
	security.InstallCipherPlugin("aes", new)
}

func new() security2.Cipher {
	cipher, err := goplugin.LookUpSymbolFromPlugin(cipherPlugin, "Cipher")
	if err != nil {
		if os.IsNotExist(err) {
			qlog.Errorf("%s not found", cipherPlugin)
		} else {
			qlog.Errorf("Load %s failed, err [%s]", cipherPlugin, err.Error())
		}
		return nil
	}
	cipherInstance, ok := cipher.(Cipher)
	if !ok {
		qlog.Infof("E: Expecting Cipher interface, but got something else.")
		return nil
	}
	cipherInstance.Init()
	return &HWAESCipher{
		gcryptoEngine: cipherInstance,
	}
}

//Encrypt is method used for encryption
func (ac *HWAESCipher) Encrypt(src string) (string, error) {
	return ac.gcryptoEngine.Encrypt(src)
}

//Decrypt is method used for decryption
func (ac *HWAESCipher) Decrypt(src string) (string, error) {
	return ac.gcryptoEngine.Decrypt(src)
}
