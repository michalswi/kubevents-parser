**POC**

Controller is watching ServiceAccount (SA) events in specific namespace (by default in `default` one). If Operator creates in the default namespace SA with suffix `-dev` for example `mytestsa-dev`, developer who is using this SA is allowed to have full access to already existing namespace `namespace-dev` dedicated for such SA. It's possible because under the hood controller will setup rbac to allow access to this namespace. If you want to limit the access you can change that in function `setupRbac` for example from full access to list and watch pods etc. 

If SA was removed from the default namespace, rbac would be removed too.


TODO  
- if SA created with suffix `-test` in default namespace should be moved to `namespace-test`