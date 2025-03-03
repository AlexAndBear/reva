// Copyright 2018-2023 CERN
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// In applying this license, CERN does not waive the privileges and immunities
// granted to it by virtue of its status as an Intergovernmental Organization
// or submit itself to any jurisdiction.

package data

// Storage defines the interface for operators and accounts storages.
type Storage interface {
	// ReadOperators reads all stored operators into the given data object.
	ReadOperators() (*Operators, error)
	// WriteOperators writes all stored operators from the given data object.
	WriteOperators(ops *Operators) error

	// OperatorAdded is called when an operator has been added.
	OperatorAdded(op *Operator)
	// OperatorUpdated is called when an operator has been updated.
	OperatorUpdated(op *Operator)
	// OperatorRemoved is called when an operator has been removed.
	OperatorRemoved(op *Operator)

	// ReadAccounts reads all stored accounts into the given data object.
	ReadAccounts() (*Accounts, error)
	// WriteAccounts writes all stored accounts from the given data object.
	WriteAccounts(accounts *Accounts) error

	// AccountAdded is called when an account has been added.
	AccountAdded(account *Account)
	// AccountUpdated is called when an account has been updated.
	AccountUpdated(account *Account)
	// AccountRemoved is called when an account has been removed.
	AccountRemoved(account *Account)
}
