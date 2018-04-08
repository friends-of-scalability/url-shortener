This is called by the Makefile through the run-k8s target which will setup a local kubernetes cluster.
`make run-k8s`

However, if you want to deploy the helm chart manually you would need
a valid kubectl context and helm installed in your system.

First you need to init helm

`helm init --upgrade`

Then package the chart using
`helm package -u ${CHARTS_PATH}/url-shortener -d ${CHARTS_PATH}`

And later install it using
`helm upgrade -i ${RELEASE_NAME} $(find ${CHARTS_PATH} -maxdepth 1 -name "*.tgz")`
