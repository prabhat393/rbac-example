# Prototype RBAC using resource role

This is a prototype implemenation of RBAC for testing purposes. 

## RBAC with user roles and resource roles

Originally, we wanted to use RBAC with user roles and resource roles ([sample model](https://github.com/casbin/casbin/blob/master/examples/rbac_with_resource_roles_model.conf) / [sample policy](https://github.com/casbin/casbin/blob/master/examples/rbac_with_resource_roles_policy.csv) ).

For example:
```
p, free, exports, GET
p, free_to_paid, exports, GET
p, unlimited, exports, GET
g2, /v1/exports/download/:namespace/:project, exports
g2, /v1/exports/meta/*, exports
```



 However, this turned out to be difficult as our resource/API need **object wildcard matching** as well as **group checking**.
 
For example,
Request: `user1` belonging to group `free` accessing `/v1/exports/download/0/enwiki`.

## RBAC with transitive user roles

Instead, we have implemented RBAC with transitive user roles ([sample policy](https://github.com/casbin/casbin/blob/master/examples/rbac_with_resource_roles_policy.csv)) (since our user do not have wildcards). This worked.

```
p, exports, /v1/exports/download/:namespace/:project, GET
p, exports, /v1/exports/meta/*, GET
g, free, exports
g, free_to_paid, exports
g, unlimited, exports
```


