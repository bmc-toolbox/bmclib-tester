#### bmclib tester

The bmclib-tester can test bmclib [features](https://github.com/bmc-toolbox/bmclib/blob/main/providers/providers.go#L5) on one more devices in parallel.

The test results are returned in the JSON format.

### build

```
go build .
```

### run

The run command accepts a `--hardware` flag which specifies the YAML file listing the hardware to to run tests on, the `--tests` flag specifies the list of bmclib features to test.

Checkout the sample test files here [hardware.yaml](samples/hardware.yaml)

The example below tests the `powerstate` feature using the `ipmi` protocol.

cmd
```
./bmclib-tester run --timeout 30s \
                    --hardware ./samples/hardware.yaml \
                    --tests ./samples/tests-ipmi.yaml`
```

output
```json
2023/05/03 07:29:45 waiting for tests to complete...
[
 {
  "Vendor": "Supermicro",
  "Model": "SYS-510T-MR",
  "Name": "c",
  "BMCIP": "192.168.1.1",
  "Results": [
   {
    "Feature": "powerstate",
    "Protocol": "ipmi",
    "Output": "Chassis Power is on\n",
    "Error": "",
    "Succeeded": true,
    "Runtime": "1ns"
   }
  ]
 },
 {
  "Vendor": "ASRockRack",
  "Model": "EPYC 7402P",
  "Name": "a",
  "BMCIP": "192.168.1.4",
  "Results": [
   {
    "Feature": "powerstate",
    "Protocol": "ipmi",
    "Output": "Chassis Power is on\n",
    "Error": "",
    "Succeeded": true,
    "Runtime": "1ns"
   }
  ]
 }
]
```