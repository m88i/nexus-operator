## Version 0.4.0

### Enhancements

- #161 - Nexus Operator is now cluster-scoped, meaning that you can install the operator and the Nexus CRs in separated namespaces. See [Operator Scopes](https://sdk.operatorframework.io/docs/building-operators/golang/operator-scope/) for more information.

### Bug Fixes
- #157 - When installing the operator and the Nexus CR in different namespaces, the server operations performed in the Nexus API by the operator won't work.