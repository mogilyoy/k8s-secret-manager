
kubectl create secret generic api-jwt-secret \
  --from-literal=JWT_SECRET=super-dev-secret-change-me \
  -n default
команду создания api-jwt-secret (или локального yaml);
make deploy → кластер готов;
kubectl apply -f config/samples/secrets_v1alpha1_secretclaim.yaml как пример использования.​

kubectl port-forward svc/api-server 8080:8080