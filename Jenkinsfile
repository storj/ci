pipeline {
    agent {
        dockerfile {
            filename 'Dockerfile'
            args '-u root:root --cap-add SYS_PTRACE -v "/tmp/gomod":/go/pkg/mod'
            label 'main'
        }
    }
    options {
          timeout(time: 26, unit: 'MINUTES')
    }
    environment {
        NPM_CONFIG_CACHE = '/tmp/npm/cache'
    }
    stages {
        stage('Build') {
            steps {
                checkout scm

                // ensure that services can start
                sh 'service postgresql start'

                sh 'cockroach start --insecure --store=\'/tmp/crdb\' --listen-addr=localhost:26257 --http-addr=localhost:8080 --join=localhost:26257 --background'
                sh 'cockroach init --insecure --host=localhost:26257'
            }
        }

        stage('Lint') {
            steps {
                sh 'check-copyright'
                sh 'check-copyright'
                sh 'check-large-files'
                sh 'check-imports ./...'
                sh 'check-peer-constraints'
                sh 'storj-protobuf --protoc=$HOME/protoc/bin/protoc lint'
                sh 'storj-protobuf --protoc=$HOME/protoc/bin/protoc check-lock'
                sh 'check-atomic-align ./...'
                sh 'check-errs ./...'
                sh 'staticcheck ./...'
                sh 'golangci-lint --config /go/ci/.golangci.yml -j=2 run'
            }
        }
    }

    post {
        always {
            sh "chmod -R 777 ." // ensure Jenkins agent can delete the working directory
            deleteDir()
        }
    }
}
