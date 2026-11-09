# Changelog

## [1.1.0](https://github.com/create2-labs/cafe-contracts/compare/v1.0.0...v1.1.0) (2026-11-09)


### ⚠ BREAKING CHANGES

* **contracts:** observation/address/v1 validation rules change: (1) chains.key_exposed is true on a delegated eoa chain whatever its nonce, instead of nonce > 0; (2) chains.nonce is required on every chain without contract code that carries no delegation_target, instead of optional; (3) related_addresses accepts a delegation_target on a not_deployed chain of a contract result, instead of eoa only; (4) contract_with_eoa_activity is also required by a not_deployed chain with a delegation_target. Results valid under 1.0.0 may be rejected, and the reverse.

### Features

* **contracts:** expose the key on a delegation designator and require a read nonce ([486356b](https://github.com/create2-labs/cafe-contracts/commit/486356b639a84c2c7ac9941701739e1922a60bd5))

## [1.0.0](https://github.com/create2-labs/cafe-contracts/compare/v0.0.1...v1.0.0) (2026-11-06)


### ⚠ BREAKING CHANGES

* **contracts:** require chain_ids in the wallet scan request and keep eoa and unknown account kinds

### Features

* **address:** add canonical EVM address helpers and EIP-55 renderer ([64382b9](https://github.com/create2-labs/cafe-contracts/commit/64382b97837147f4c16ba1702f32ab5014e4be6a))
* **address:** add shared canonical address utilities ([6a81115](https://github.com/create2-labs/cafe-contracts/commit/6a8111564fbb8b7953ab69d4584bdb14b1d02bde))
* **address:** add shared canonical address utilities ([618a648](https://github.com/create2-labs/cafe-contracts/commit/618a64886caff9e80826e21641c69f0b1617d8db))
* **address:** add shared canonical address utilities ([229c8ad](https://github.com/create2-labs/cafe-contracts/commit/229c8add03f08c982f096d8bb075723c3b61f734))
* **cafenatsv01:** assessment NATS wire v0.2 without selection_request ([9c1a484](https://github.com/create2-labs/cafe-contracts/commit/9c1a48486f0e1d70ae6250da8743dd771a243cc5))
* **cafenatsv01:** replace key_rotation_required with key_rotation_model ([09359e7](https://github.com/create2-labs/cafe-contracts/commit/09359e72660dc5b8704d34ed3e88576c7a1d0300))
* **cafenatsv01:** replace key_rotation_required with key_rotation_model ([01fea01](https://github.com/create2-labs/cafe-contracts/commit/01fea0174749dc8f4871203e22f2fc97be4a60c9))
* **contracts:** add enrichment.completed with per-chain outcomes ([4afb455](https://github.com/create2-labs/cafe-contracts/commit/4afb45587a07babd82254c266282974741225763))
* **contracts:** add enrichment.completed with per-chain outcomes ([6cfd583](https://github.com/create2-labs/cafe-contracts/commit/6cfd58346b5db9dfbef8ae1612a03a8558e76097))
* **contracts:** add enrichment.started and enrichment.failed ([a81b47d](https://github.com/create2-labs/cafe-contracts/commit/a81b47d9c3835d7a9108b39dfb399ba62b358f46))
* **contracts:** add enrichment.started and enrichment.failed ([63e1347](https://github.com/create2-labs/cafe-contracts/commit/63e1347fdbbb9759d747de760ae06d2126510d5f))
* **contracts:** add explicit assessment request and CPM/remediation event contracts v0.1 ([59f7747](https://github.com/create2-labs/cafe-contracts/commit/59f7747c6198661f18003575c57f3d4fe8943821))
* **contracts:** add explicit assessment request and policy/remediation event contracts v0.1 ([424877e](https://github.com/create2-labs/cafe-contracts/commit/424877e1778aa16eb88e5191555abe73520b0e1d))
* **contracts:** add heartbeat and persistence.ready wire contracts ([124da48](https://github.com/create2-labs/cafe-contracts/commit/124da48062612951535210d9b6346534f93752c2))
* **contracts:** add heartbeat and persistence.ready wire contracts ([53c2818](https://github.com/create2-labs/cafe-contracts/commit/53c2818bdaf21d25ec1700cae3be820cbedb2051))
* **contracts:** add indexer page and scanner presence wire contracts ([3534fc5](https://github.com/create2-labs/cafe-contracts/commit/3534fc5d96a90c850af57efa9a26999ce0df0b80))
* **contracts:** add indexer page and scanner presence wire contracts ([721c54f](https://github.com/create2-labs/cafe-contracts/commit/721c54f1d69fb1df2fcad43aa8a9565b977a9b12))
* **contracts:** add observation/wallet/v01 normative wire contract ([96eec80](https://github.com/create2-labs/cafe-contracts/commit/96eec8063ae4569ffd5c4fc8dc080c9573c7ab1a))
* **contracts:** add observation/wallet/v01 normative wire contract ([69fa7f7](https://github.com/create2-labs/cafe-contracts/commit/69fa7f7bb0fc7375c8339a4c39536b4569bd666b))
* **contracts:** add related addresses, anomalies, and EOA facts to observation/address/v1 ([d445669](https://github.com/create2-labs/cafe-contracts/commit/d445669ed1609d187ffbf60bef3f804103266f9e))
* **contracts:** add related addresses, anomalies, and EOA facts to observation/address/v1 ([9f25594](https://github.com/create2-labs/cafe-contracts/commit/9f25594e3e154e6cc71ccd5b80b7f28580dc4147))
* **contracts:** add scan.started, scan.failed, and scan.ready ([c3d480e](https://github.com/create2-labs/cafe-contracts/commit/c3d480e06157bf66332eaa30942a7c3ca11ed302))
* **contracts:** add scan.started, scan.failed, and scan.ready ([06f7ff7](https://github.com/create2-labs/cafe-contracts/commit/06f7ff7e9c7ca16412a466ca3c84400d231822de))
* **contracts:** add the contract creation gateway request ([a32d9c5](https://github.com/create2-labs/cafe-contracts/commit/a32d9c5f87290b5abd6601af479857867c763523))
* **contracts:** add the contract creation gateway request ([eddf155](https://github.com/create2-labs/cafe-contracts/commit/eddf155f97f2b659f6a0511bb6ce2f0aa9bedc48))
* **contracts:** add the enrichment request contract ([36d41dd](https://github.com/create2-labs/cafe-contracts/commit/36d41dd0c87c589d70b2dbb6976aa86a18b4bc36))
* **contracts:** add the enrichment request contract ([eb550a5](https://github.com/create2-labs/cafe-contracts/commit/eb550a5fed1c1555e64c53183fb34a4dad3a7047))
* **contracts:** add the heartbeat subjects of enrichment workers ([6a26a0f](https://github.com/create2-labs/cafe-contracts/commit/6a26a0fbdd7ca19d2d9a9b525109d849fb4e6968))
* **contracts:** add the heartbeat subjects of enrichment workers ([35b4c80](https://github.com/create2-labs/cafe-contracts/commit/35b4c807b023de734812b20b276101f5943dd6f2))
* **contracts:** add the observation/address/v1 qualification result ([769bb24](https://github.com/create2-labs/cafe-contracts/commit/769bb24b84ca579a51550dc2ea9a518527b5289c))
* **contracts:** add the observation/address/v1 qualification result ([5040954](https://github.com/create2-labs/cafe-contracts/commit/504095408445e17b493d38459b98800eba395118))
* **contracts:** add the scan request wire contract ([a3280bb](https://github.com/create2-labs/cafe-contracts/commit/a3280bb67073e558193a2cc858d4dcee66eb0d7c))
* **contracts:** add the scan request wire contract ([0e88bd6](https://github.com/create2-labs/cafe-contracts/commit/0e88bd6d1b8cde0764f909dc681a53952e980ad8))
* **contracts:** add the scan.completed wire contract ([e8ea9ce](https://github.com/create2-labs/cafe-contracts/commit/e8ea9cee618d5c08a95ada475294d3f1e865ca0a))
* **contracts:** add the scan.completed wire contract ([6e8f6e8](https://github.com/create2-labs/cafe-contracts/commit/6e8f6e809106ea61e04c88144ec4bee88e7cc1d0))
* **contracts:** decode the contract creation gateway response ([c718ef8](https://github.com/create2-labs/cafe-contracts/commit/c718ef859a5c7b26f3174459859413e82d68d84b))
* **contracts:** decode the contract creation gateway response ([6a738d2](https://github.com/create2-labs/cafe-contracts/commit/6a738d2d3ff4b76f75eb28477c69869974f50014))
* **contracts:** require chain_ids in the wallet scan request and keep eoa and unknown account kinds ([22bcc3a](https://github.com/create2-labs/cafe-contracts/commit/22bcc3a3bd29b5e5ad40810649431d63edb29f23))
* **contracts:** validate the contract creation gateway response and lowercase the request address ([ffa2394](https://github.com/create2-labs/cafe-contracts/commit/ffa2394b3ec46b73a5d866142589da259d252c68))
* **contracts:** validate the contract creation gateway response and lowercase the request address ([090be42](https://github.com/create2-labs/cafe-contracts/commit/090be426af1c4652fd7db784dffca7fbdb436197))
* **cwr0-1:** introduce eventenvelope/v01 header contract and tests ([d00789b](https://github.com/create2-labs/cafe-contracts/commit/d00789b25695fa251971979b8217bef991daa51e))
* upgrade to Go 1.26.9 ([5dd4daa](https://github.com/create2-labs/cafe-contracts/commit/5dd4daa04c1e34f0dea4cd5c027d30c7fd01d370))
* upgrade to Go 1.26.9 ([ae33bfe](https://github.com/create2-labs/cafe-contracts/commit/ae33bfe3f7bfcca89822a227c0746399a4eec141))
* upgrade to Go 1.27.2 ([5709708](https://github.com/create2-labs/cafe-contracts/commit/5709708a5f807557ea10db3ec0502864ff359239))


### Bug Fixes

* avoid gosec G602 false positive in ToChecksumEIP55 ([b39c6dc](https://github.com/create2-labs/cafe-contracts/commit/b39c6dcf2dbc2690b470f6c13efdb0b69710c830))
* errors uses constants.go ([9bfe4a4](https://github.com/create2-labs/cafe-contracts/commit/9bfe4a451e942ae1ea5c8628b81a8b25cff243f0))
* upgrade to Go 1.27.3 ([8790849](https://github.com/create2-labs/cafe-contracts/commit/8790849982dddc76b1a943afa8420756ab003664))
* upgrade to Go 1.27.3 ([662e6ab](https://github.com/create2-labs/cafe-contracts/commit/662e6abfa0c87e3044601a12603addba8a936fc2))
* upgraded to go 1.16.4 ([21574f8](https://github.com/create2-labs/cafe-contracts/commit/21574f855de11f1988b2fffc96947d5dc31695a3))
