# Apps for Mobius

[Mobius](https://github.com/mobius-scheduler/mobius) is a scheduling system for shared mobility platforms. This repository contains implementations for different apps (customer code) that we deploy atop Mobius. Mobius supports apps with diverse spatiotemporal requirements, streaming requests, and reactive preferences. All apps use the [`InterestMap`](https://github.com/mobius-scheduler/mobius/blob/main/common/interestmap.go) interface that Mobius exposes. Some of the apps we implemented include:
* `aqi/`: Mapping air quality using a Gaussian Process model, from aerial PM2.5 measurements.
* `iperf/`: Profiling cellular throughput in the air by running iPerf at prescribed measurement locations.
* `parking/`: Regularly parking space occupancy from aerial imagery.
* `roof/`: Image roofs in a residential area.
* `traffic/`: Monitor real-time road traffic congestion from aerial videos.
* `lyft/`: Receive Lyft ride requests by originating neighborhood in a metro area.

Note that these implementations require traces containing ground-truth measurements, in order to run within Mobius' trace-driven emulation framework. However, these app implementations are compatible with a real-time system where measurements are gathered or requests arrive at runtime.

For more details on these apps, check out our [paper](https://web.mit.edu/arjunvb/pubs/mobius-mobisys21-paper.pdf).
