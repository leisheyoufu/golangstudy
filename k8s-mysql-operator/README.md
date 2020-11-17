## Install operator-sdk
RELEASE_VERSION=v1.1.0
curl -LO https://github.com/operator-framework/operator-sdk/releases/download/${RELEASE_VERSION}/operator-sdk-${RELEASE_VERSION}-x86_64-apple-darwin
chmod +x operator-sdk-v1.1.0-x86_64-apple-darwin
mv operator-sdk-v1.1.0-x86_64-apple-darwin /usr/local/bin/operator-sdk

## Init project
operator-sdk init --domain=example.com --repo=github.com/example-inc/mysql-operator
operator-sdk create api --group db --version v1 --kind Mysql --resource=true --controller=true

## types
make generate  // deepcopy
make manifests // crd file

kubectl apply -f config/crd/bases/db.example.com_mysqls.yaml
make run // start controller outside k8s cluster
kubectl apply -f config/samples/db_v1_mysql.yaml

## Reference

[operator-sdk-quickstart](https://sdk.operatorframework.io/docs/building-operators/golang/quickstart/)
[operator step by step](https://www.katacoda.com/openshift/courses/operatorframework/go-operator-podset)
https://github.com/ica10888/multi-tenancy-operator
[k8s list watch](https://www.youtube.com/watch?v=PLSDvFjR9HY)
[kruise](https://github.com/openkruise/kruise)
