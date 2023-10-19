package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/p2p/pex"
)

const (
	// routabilityStrict is a hard-coded config value for the address book.
	// See https://github.com/celestiaorg/celestia-core/blob/793ece9bbd732aec3e09018e37dc31f4bfe122d9/config/config.go#L540-L542
	routabilityStrict = true
)

func addrbookCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addrbook peers.txt addrbook.json",
		Short: "Convert a list of peers into an address book",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFile := args[0]
			outputFile := args[1]

			data, err := os.ReadFile(inputFile)
			if err != nil {
				return err
			}
			lines := strings.Split(string(data), "\n")

			book := pex.NewAddrBook(outputFile, routabilityStrict)
			for _, line := range lines {
				if line == "" {
					continue
				}
				address, err := p2p.NewNetAddressString(line)
				if err != nil {
					return err
				}
				err = book.AddAddress(address, address)
				if err != nil {
					return err
				}
			}

			book.Save()
			fmt.Printf("Converted %s into %s\n", inputFile, outputFile)
			return nil
		},
	}

	return cmd
}
