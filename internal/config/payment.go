package config

import (
	"github.com/stellar/go/keypair"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type PaymentConfig struct {
	SourceSigner  *keypair.Full
	SourceAddress *keypair.FromAddress
}

func (c *config) PaymentConfig() PaymentConfig {
	var result struct {
		SourceSigner  string `fig:"source_signer"`
		SourceAddress string `fig:"source_address"`
	}

	err := figure.
		Out(&result).
		With(figure.BaseHooks).
		From(kv.MustGetStringMap(c.getter, "payment")).
		Please()
	if err != nil {
		panic(errors.Wrap(err, "failed to figure out payment"))
	}

	c.paymentConfig = PaymentConfig{
		SourceAddress: keypair.MustParse(result.SourceAddress).(*keypair.FromAddress),
		SourceSigner:  keypair.MustParse(result.SourceSigner).(*keypair.Full),
	}
	return c.paymentConfig
}
