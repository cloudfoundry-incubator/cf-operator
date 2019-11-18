## Use Cases

- [Use Cases](#use-cases)
  - [exstatefulset_configs.yaml](#exstatefulsetconfigsyaml)
  - [exstatefulset_configs_updated.yaml](#exstatefulsetconfigsupdatedyaml)
  - [exstatefulset_azs.yaml](#exstatefulsetazsyaml)
  - [exstatefulset_pvcs.yaml](#exstatefulsetpvcsyaml)

### exstatefulset_configs.yaml

This creates a `StatefulSet` with two `Pods`.

### exstatefulset_configs_updated.yaml

This is a copy of `exstatefulset_configs.yaml`, with one config value changed. 

When applied on top using `kubectl`, this exemplifies the automatic updating of the `Pods` with a new value for the `SPECIAL_KEY` environment variable.

### exstatefulset_azs.yaml

This creates 4 `Pods` - 2 in one zone and 2 in another zone.

### exstatefulset_pvcs.yaml

This creates `Statefulset Pods` with `Persistent Volumes Claims` attached to each `Pod`. The created `Persistent Volume Claims` get re-attached to the new versions of StatefulSet Pods when the QuarksStatefulSet is updated.