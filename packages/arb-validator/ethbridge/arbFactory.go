/*
 * Copyright 2020, Offchain Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ethbridge

import (
	"context"
	"math/big"

	"github.com/offchainlabs/arbitrum/packages/arb-validator/structures"

	"github.com/offchainlabs/arbitrum/packages/arb-validator/ethbridge/arbfactory"
	errors2 "github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ArbFactory struct {
	contract *arbfactory.ArbFactory
	client   *ethclient.Client
	auth     *bind.TransactOpts
}

func NewArbFactory(address common.Address, client *ethclient.Client, auth *bind.TransactOpts) (*ArbFactory, error) {
	vmCreatorContract, err := arbfactory.NewArbFactory(address, client)
	if err != nil {
		return nil, errors2.Wrap(err, "Failed to connect to ArbFactory")
	}
	return &ArbFactory{vmCreatorContract, client, auth}, nil
}

func (con *ArbFactory) GlobalPendingInboxAddress() (common.Address, error) {
	return con.contract.GlobalInboxAddress(nil)
}

func (con *ArbFactory) ChallengeFactoryAddress() (common.Address, error) {
	return con.contract.ChallengeFactoryAddress(nil)
}

func (con *ArbFactory) CreateRollup(
	ctx context.Context,
	vmState [32]byte,
	params structures.ChainParams,
	owner common.Address,
) (common.Address, error) {
	con.auth.Context = ctx
	tx, err := con.contract.CreateRollup(
		con.auth,
		vmState,
		params.GracePeriod.Val,
		new(big.Int).SetUint64(params.ArbGasSpeedLimitPerTick),
		params.MaxExecutionSteps,
		params.StakeRequirement,
		owner,
	)
	if err != nil {
		return common.Address{}, errors2.Wrap(err, "Failed to call to ChainFactory.CreateChain")
	}
	receipt, err := WaitForReceiptWithResults(ctx, con.client, con.auth.From, tx, "CreateChain")
	if err != nil {
		return common.Address{}, err
	}
	if len(receipt.Logs) != 1 {
		return common.Address{}, errors2.New("Wrong receipt count")
	}
	event, err := con.contract.ParseRollupCreated(*receipt.Logs[0])
	if err != nil {
		return common.Address{}, err
	}
	return event.VmAddress, nil
}

type ArbFactoryWatcher struct {
	contract *arbfactory.ArbFactory
	client   *ethclient.Client
}

func NewArbFactoryWatcher(address common.Address, client *ethclient.Client) (*ArbFactoryWatcher, error) {
	vmCreatorContract, err := arbfactory.NewArbFactory(address, client)
	if err != nil {
		return nil, errors2.Wrap(err, "Failed to connect to ArbFactory")
	}
	return &ArbFactoryWatcher{vmCreatorContract, client}, nil
}

func (con *ArbFactoryWatcher) GlobalPendingInboxAddress() (common.Address, error) {
	return con.contract.GlobalInboxAddress(nil)
}

func (con *ArbFactoryWatcher) ChallengeFactoryAddress() (common.Address, error) {
	return con.contract.ChallengeFactoryAddress(nil)
}
