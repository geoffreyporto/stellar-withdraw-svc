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
	MaxBaseFee    uint32
}

func (c *config) PaymentConfig() PaymentConfig {
	var result struct {
		SourceSigner  string `fig:"source_signer"`
		SourceAddress string `fig:"source_address"`
		MaxBaseFee    uint32`fig:"max_base_fee"`
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
		MaxBaseFee:    result.MaxBaseFee,
	}
	return c.paymentConfig
}
