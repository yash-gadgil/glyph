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

k8s_resource('rabbitmq', port_forwards=['5672:5672', '15672:15672'])
k8s_resource('order-book', resource_deps=['rabbitmq'])
k8s_resource('userdb', port_forwards=['5432:5432'])
k8s_resource('user', resource_deps=['userdb', 'rabbitmq'], port_forwards=['50053:50053'])
k8s_resource('orderdb', port_forwards=['5433:5432'])
k8s_resource('order', resource_deps=['orderdb', 'rabbitmq', 'order-book', 'user'], port_forwards=['50055:50055'])
