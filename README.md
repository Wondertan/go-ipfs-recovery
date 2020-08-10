[![Build](https://img.shields.io/circleci/build/github/Wondertan/go-ipfs-recovery.svg?style=svg)](https://img.shields.io/circleci/build/github/Wondertan/go-ipfs-recovery)
[![Go Report Card](https://goreportcard.com/badge/github.com/Wondertan/go-ipfs-recovery)](https://goreportcard.com/report/github.com/Wondertan/go-ipfs-recovery)
[![License](https://img.shields.io/github/license/Wondertan/go-ipfs-recovery.svg?maxAge=2592000)](https://github.com/Wondertan/go-ipfs-recovery/blob/master/LICENSE)
[![codecov](https://codecov.io/gh/Wondertan/go-ipfs-recovery/branch/master/graph/badge.svg)](https://codecov.io/gh/Wondertan/go-ipfs-recovery)

# go-ipfs-recovery

> The project was originally started as a submission for the [HackFS](https://hackfs.com/) hackaton.

Golang implementation for IPFS Recovery.

## Table of Contents

## Related Work

### IPFS Fork
As Recovery follows IPFS ecosystem modularity best practises, 
it [fork](https://github.com/Wondertan/go-ipfs/tree/recovery) is integrated in a just a few small changes. 

First, it covers DAG sessions
with custom NodeGetter that can recover nodes on the fly if content is requested but not found.
Furthermore, fork adds additional functionality to IPFS CLI extending it with `recovery` command group. Currently, 
it is only capable for encoding DAGs with Reed Solomon recoverability using `encode`, but later CLI will be extended with
full featured management for Recovery, like re-encoding, manual recovery and algorithm choices.

### Testground plans
IPFS ecosystem recently launched new project aimed to test p2p system on large scale to simulate real world behavior. 
Using it for benchmarking and [testing Recovery](https://github.com/avalonche/bitswap-recovery) is a must, as it goals to improve IPFS protocol.
