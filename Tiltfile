
docker_build(
    'glyph/gateway', 
    '.', 
    dockerfile='deployments/docker/gateway.Dockerfile',
    build_args={'BUILDKIT_INLINE_CACHE': '1'},
    extra_tag='glyph/gateway:latest',
    pull=False  
)

docker_build(
    'glyph/auth', 
    '.', 
    dockerfile='deployments/docker/auth.Dockerfile',
    build_args={'BUILDKIT_INLINE_CACHE': '1'},
    extra_tag='glyph/auth:latest',
    pull=False  
)

docker_build(
    'glyph/mrktdata', 
    '.', 
    dockerfile='deployments/docker/mrktdata.Dockerfile',
    build_args={'BUILDKIT_INLINE_CACHE': '1'},
    extra_tag='glyph/mrktdata:latest',
    pull=False  
)

docker_build(
    'glyph/user', 
    '.', 
    dockerfile='deployments/docker/user.Dockerfile',
    build_args={'BUILDKIT_INLINE_CACHE': '1'},
    extra_tag='glyph/user:latest',
    pull=False  
)


k8s_yaml('deployments/k8s/namespace.yaml')

k8s_yaml([
    'deployments/k8s/auth/auth-config.yaml',
    'deployments/k8s/auth/auth-deployment.yaml',
    'deployments/k8s/auth/auth-secrets.yaml',
    'deployments/k8s/auth/auth-service.yaml'
])

k8s_yaml([
    'deployments/k8s/mrktdata/mrktdata-config.yaml',
    'deployments/k8s/mrktdata/mrktdata-deployment.yaml',
    'deployments/k8s/mrktdata/mrktdata-secrets.yaml',
    'deployments/k8s/mrktdata/mrktdata-service.yaml'
])

k8s_yaml([
    'deployments/k8s/gateway/gateway-config.yaml',
    'deployments/k8s/gateway/gateway-deployment.yaml',
    'deployments/k8s/gateway/gateway-service.yaml'
])

k8s_yaml([
    'deployments/k8s/user/user-config.yaml',
    'deployments/k8s/user/user-deployment.yaml',
    'deployments/k8s/user/user-service.yaml',
    'deployments/k8s/user/user-migrations.yaml'
])

k8s_yaml([
    'deployments/k8s/userdb/userdb-pvc.yaml',
    'deployments/k8s/userdb/userdb-config.yaml',
    'deployments/k8s/userdb/userdb-service.yaml',
    'deployments/k8s/userdb/userdb-deployment.yaml'
])

k8s_yaml([
    'deployments/k8s/authcache/authcache-deployment.yaml',
    'deployments/k8s/authcache/authcache.yaml'
])

k8s_resource('user-migrations', resource_deps=['userdb'])
k8s_resource('user', resource_deps=['user-migrations'])

k8s_resource('auth', resource_deps=['auth-cache'])

k8s_resource('gateway', port_forwards=3000)
k8s_resource('mrktdata', port_forwards=50052)