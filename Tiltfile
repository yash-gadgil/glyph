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

k8s_resource('rabbitmq', port_forwards=['5672:5672', '15672:15672'])
k8s_resource('order-book', resource_deps=['rabbitmq'])
