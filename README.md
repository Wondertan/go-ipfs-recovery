[![Build](https://img.shields.io/circleci/build/github/Wondertan/go-ipfs-recovery.svg?style=svg)](https://img.shields.io/circleci/build/github/Wondertan/go-ipfs-recovery)
[![Go Report Card](https://goreportcard.com/badge/github.com/Wondertan/go-ipfs-recovery)](https://goreportcard.com/report/github.com/Wondertan/go-ipfs-recovery)
[![License](https://img.shields.io/github/license/Wondertan/go-ipfs-recovery.svg?maxAge=2592000)](https://github.com/Wondertan/go-ipfs-recovery/blob/master/LICENSE)
[![codecov](https://codecov.io/gh/Wondertan/go-ipfs-recovery/branch/master/graph/badge.svg)](https://codecov.io/gh/Wondertan/go-ipfs-recovery)

# IPFS Recovery

> The project was originally started as a [submission](https://hack.ethglobal.co/hackfs/teams/recBTnbaJZ9h8JJUE/rec909D6romwHglDV) 
> for the [HackFS](https://hackfs.com/) hackaton. Also check-out our [presentation](https://drive.google.com/file/d/1wyO7Zt5gAXuQUOh2Nlf_lhEjdZO2MelQ/view).

Building a way for content to persist permanently despite any damage to data and the network by  
bringing data recovery algorithms into IPFS protocol.

## Table of Contents

- [Background](#background)
- [Implementation](#implementation)
- [Algorithms](#algorithms)
- [Related Work](#related-work)
- [Future Work](#future-work)
- [Tryout](#tryout)
- [Contributors](#contributors)
- [License](#license)

## Background

The IPFS project, in its core, tries to upgrade the Internet to make it better in multiple ways. One of the goals is to 
make the Web permanent. This IPFS characteristic is very promising, but still not in the state of the art form and 
requires more RnD in that vector. On the other side, Computer Science, for many years of existence, did multiple 
inventions related to data and the ways to make it persistent against multiple data-loss factors within mostly 
centralized systems. On the way to permanency, those inventions can apply to IPFS protocol taking the most out of them, 
and then newer innovations might take place instead after gathering all the experience. Further work in this avenue can 
also ensure integrity even in doomsday scenarios where large portions of the network can go down while it is still possible 
to recover that content. Though there are multiple discussions in the IPFS ecosystem regarding the data persevering mechanisms, 
like erasure codings, none of them were actually implemented.

The IPFS Recovery project brings data recovery algorithms into IPFS with the above aim. It does so by creating new IPLD 
data structures able to do self-recovery in case some of the nodes are lost and can't be found on the network due to 
node churn, network issues, or physical storage damage.

## Implementation

The Recovery currently points to the main IPFS implementation in Golang and follows all its development guidelines and
best practices. The Golang Recovery implementation is a fully modular library with clean API boundaries that aims to 
provide convenient use and excellent abstraction for all current and future implementations.

## Algorithms

### Reed-Solomon

For the initial version, the project started with industry-standard [Reed-Solomon](https://www2.cs.duke.edu/courses/spring10/cps296.3/rs_scribe.pdf) 
coding.

### Alpha Entanglements

As a next step, novel [Alpha Entanglements](https://arxiv.org/pdf/1810.02974.pdf) schema have been chosen. It provides
better performance and higher recovery ratio comparing with the former algorithm. In particular, entaglements are
interesting as they provide the ability to create self-healing networks.

## Related Work

### IPFS Fork

As Recovery follows IPFS ecosystem modularity best practises, 
it [fork](https://github.com/Wondertan/go-ipfs/tree/recovery) is integrated in a just a few small changes. 

First, it covers DAG sessions
with custom NodeGetter that can recover nodes on the fly if content is requested but not found.
Furthermore, fork adds additional functionality to IPFS CLI extending it with `recovery` command group. Currently, 
it is only capable for encoding DAGs with Reed Solomon recoverability using `encode`, but later CLI will be extended with
full featured management for Recovery, like re-encoding, manual recovery and algorithm choices.

### Testground Plans

IPFS ecosystem recently launched new project aimed to test p2p system on large scale to simulate real world behavior. 
Using it for benchmarking and [testing Recovery](https://github.com/avalonche/bitswap-recovery) is a must, 
as it goals to improve IPFS protocol.

## Future Work

- Upgrade to more complex Alpha Entanglement parity lattice to reach better
  performance.
  
- IPLD specs formalization through active discussions and feedback processing.

- Implementations for latest `go` IPLD version and for `js` as well.

- Extensive Testground simulation to gather real-world resiliency benchmarks and
  to examine various other erasure codes

## Tryout

1. Build [forked IPFS](https://github.com/Wondertan/go-ipfs/tree/recovery)
2. Encode **ANY** IPFS content: `ipfs recovery encode <path>`
3. List all the blocks encoded content consist of: `ipfs refs <enc_cid> -r`
4. Remove any random blocks yourself from the given list: `ipfs block rm ...<cid>`
5. Be amazed after seeing that it is still possible to get your content back: `ipfs get <enc_cid>`

## Contributors

- [@Wondertan](https://github.com/Wondertan)
- [@govi218](https://github.com/govi218)
- [@avalonche](https://github.com/avalonche)

## License

[MIT © Hlib Kanunnikov](https://github.com/Wondertan/go-ipfs-recovery/blob/master/LICENSE)
