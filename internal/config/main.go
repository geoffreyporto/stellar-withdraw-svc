package config

import (
	"github.com/stellar/go/clients/horizonclient"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type config struct {
	paymentConfig  PaymentConfig
	stellar        horizonclient.ClientInterface
	withdrawConfig WithdrawConfig

	getter kv.Getter
	once   comfig.Once
	Horizoner
	comfig.Logger
}

type Config interface {
	WithdrawConfig() WithdrawConfig
	PaymentConfig() PaymentConfig
	Stellar() horizonclient.ClientInterface
	Log() *logan.Entry
	Horizoner
}

func NewConfig(getter kv.Getter) Config {
	return &config{
		getter:    getter,
		Horizoner: NewHorizoner(getter),
		Logger:    comfig.NewLogger(getter, comfig.LoggerOpts{}),
	}
}

func (c *config) Stellar() horizonclient.ClientInterface {
	c.once.Do(func() interface{} {
		var result struct {
			IsTestNet bool `fig:"is_testnet"`
		}

		err := figure.
			Out(&result).
			With(figure.BaseHooks).
			From(kv.MustGetStringMap(c.getter, "stellar")).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out stellar"))
		}
		var cl horizonclient.ClientInterface
		switch result.IsTestNet {
		case true:
			cl = horizonclient.DefaultTestNetClient
		case false:
			cl = horizonclient.DefaultPublicNetClient
		}
		c.stellar = cl

		return nil
	})
	return c.stellar
}
