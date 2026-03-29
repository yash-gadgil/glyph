docker_build(
    'glyph/order_book',
    '.',
    dockerfile='deployments/docker/order_book.Dockerfile',
    only=[
        'services/order_book/',
        'proto/',
    ],
)

k8s_yaml('deployments/k8s/namespace.yaml')

k8s_yaml([
    'deployments/k8s/rabbitmq/rabbitmq-config.yaml',
    'deployments/k8s/rabbitmq/rabbitmq-deployment.yaml',
    'deployments/k8s/rabbitmq/rabbitmq-service.yaml',
])

k8s_yaml([
    'deployments/k8s/order_book/order_book-config.yaml',
    'deployments/k8s/order_book/order_book-deployment.yaml',
    'deployments/k8s/order_book/order_book-service.yaml',
])

docker_build(
    'glyph/user',
    '.',
    dockerfile='deployments/docker/user.Dockerfile',
    only=[
        'go.mod',
        'go.sum',
        'pkg/',
        'services/gen/',
        'services/user/',
    ],
)

k8s_yaml([
    'deployments/k8s/userdb/userdb-config.yaml',
    'deployments/k8s/userdb/userdb-volume.yaml',
    'deployments/k8s/userdb/userdb-pvc.yaml',
    'deployments/k8s/userdb/userdb-deployment.yaml',
    'deployments/k8s/userdb/userdb-service.yaml',
])

k8s_yaml([
    'deployments/k8s/user/user-config.yaml',
    'deployments/k8s/user/user-deployment.yaml',
    'deployments/k8s/user/user-service.yaml',
])

docker_build(
    'glyph/order',
    '.',
    dockerfile='deployments/docker/order.Dockerfile',
    only=[
        'go.mod',
        'go.sum',
        'pkg/',
        'services/gen/',
        'services/order/',
    ],
)

k8s_yaml([
    'deployments/k8s/orderdb/orderdb-config.yaml',
    'deployments/k8s/orderdb/orderdb-volume.yaml',
    'deployments/k8s/orderdb/orderdb-pvc.yaml',
    'deployments/k8s/orderdb/orderdb-deployment.yaml',
    'deployments/k8s/orderdb/orderdb-service.yaml',
])

k8s_yaml([
    'deployments/k8s/order/order-config.yaml',
    'deployments/k8s/order/order-deployment.yaml',
    'deployments/k8s/order/order-service.yaml',
])

docker_build(
    'glyph/mrktdata',
    '.',
    dockerfile='deployments/docker/mrktdata.Dockerfile',
    only=[
        'go.mod',
        'go.sum',
        'pkg/',
        'services/gen/',
        'services/mrktdata/',
    ],
)

k8s_yaml([
    'deployments/k8s/mrktdata/mrktdata-config.yaml',
    'deployments/k8s/mrktdata/mrktdata-deployment.yaml',
    'deployments/k8s/mrktdata/mrktdata-secrets.yaml',
    'deployments/k8s/mrktdata/mrktdata-service.yaml',
])

docker_build(
    'glyph/auth',
    '.',
    dockerfile='deployments/docker/auth.Dockerfile',
    only=[
        'go.mod',
        'go.sum',
        'pkg/',
        'services/gen/',
        'services/auth/',
    ],
)

k8s_yaml([
    'deployments/k8s/authcache/authcache-deployment.yaml',
    'deployments/k8s/authcache/authcache.yaml',
])

k8s_yaml([
    'deployments/k8s/auth/auth-config.yaml',
    'deployments/k8s/auth/auth-deployment.yaml',
    'deployments/k8s/auth/auth-secrets.yaml',
    'deployments/k8s/auth/auth-service.yaml',
])

docker_build(
    'glyph/inference',
    '.',
    dockerfile='deployments/docker/inference.Dockerfile',
    only=[
        'services/gen/python/',
        'services/inference/',
    ],
)

k8s_yaml([
    'deployments/k8s/inference/inference-config.yaml',
    'deployments/k8s/inference/inference-volume.yaml',
    'deployments/k8s/inference/inference-pvc.yaml',
    'deployments/k8s/inference/inference-deployment.yaml',
    'deployments/k8s/inference/inference-service.yaml',
])

docker_build(
    'glyph/advisor',
    '.',
    dockerfile='deployments/docker/advisor.Dockerfile',
    only=[
        'go.mod',
        'go.sum',
        'pkg/',
        'services/gen/',
        'services/advisor/',
    ],
)

k8s_yaml([
    'deployments/k8s/advisor/advisor-config.yaml',
    'deployments/k8s/advisor/advisor-deployment.yaml',
    'deployments/k8s/advisor/advisor-service.yaml',
])

docker_build(
    'glyph/gateway',
    '.',
    dockerfile='deployments/docker/gateway.Dockerfile',
    only=[
        'go.mod',
        'go.sum',
        'pkg/',
        'services/gen/',
        'services/gateway/',
    ],
)

k8s_yaml([
    'deployments/k8s/gateway/gateway-config.yaml',
    'deployments/k8s/gateway/gateway-deployment.yaml',
    'deployments/k8s/gateway/gateway-service.yaml',
])

k8s_resource('rabbitmq', port_forwards=['5672:5672', '15672:15672'])
k8s_resource('order-book', resource_deps=['rabbitmq'])
k8s_resource('mrktdata', resource_deps=['order-book'])
k8s_resource('auth-cache')
k8s_resource('auth', resource_deps=['auth-cache', 'user'])
k8s_resource('userdb', port_forwards=['5432:5432'])
k8s_resource('user', resource_deps=['userdb', 'rabbitmq'], port_forwards=['50053:50053'])
k8s_resource('orderdb', port_forwards=['5433:5432'])
k8s_resource('order', resource_deps=['orderdb', 'rabbitmq', 'order-book', 'user'], port_forwards=['50055:50055'])
k8s_resource('inference', port_forwards=['50057:50057'])
k8s_resource('advisor', resource_deps=['user', 'inference'], port_forwards=['50058:50058'])
k8s_resource('gateway', resource_deps=['auth', 'user', 'mrktdata', 'order', 'advisor'], port_forwards=['8080:8080'])

k8s_yaml([
    'deployments/k8s/monitoring/prometheus-config.yaml',
    'deployments/k8s/monitoring/prometheus-deployment.yaml',
    'deployments/k8s/monitoring/grafana-datasource.yaml',
    'deployments/k8s/monitoring/grafana-dashboard.yaml',
    'deployments/k8s/monitoring/grafana-deployment.yaml',
])

k8s_resource('prometheus', port_forwards=['9090:9090'])
k8s_resource('grafana', resource_deps=['prometheus'], port_forwards=['3001:3000'])
