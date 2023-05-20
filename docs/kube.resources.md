# synthetos kube resources

List resources associated with deployed kubernetes objects.

```sh
kuberium kube resources --kube.namespace=STRING --kube.path=STRING --kube.context=STRING
```
## Flags

The following flags can be used to configure this command:

| Name                                  | Environment Variable      | Default Value | Description                              |
| :-------------------------------------| :-------------------------| :-------------| :----------------------------------------|
| --[no]-kube.resources.concise \<bool> | SC_KUBE_RESOURCES_CONCISE | true          | Do not include a detailed resource list. |
