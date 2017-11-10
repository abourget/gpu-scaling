gpu-scaling
===========

Side-car container to scale a Kubernetes Deployment according to its
GPU usage (through nvidia-smi).



`gpu-scaler-reporter`
---------------------

Runs as a sidecar to your GPU-consuming pod, and reports metrics to the `gpu-scaler-ctrl`.


`cpu-scaler-ctrl`
-----------------

Scales the Deployment according to avg usage of the GPUs reported by
`gpu-scaler-reporter`, based on configurable thresholds.
