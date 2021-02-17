def repositories = [
    'https://review.dev.storj.io/storj/common',
    'https://review.dev.storj.io/storj/uplink',
    'https://review.dev.storj.io/storj/uplink-c',
    'https://review.dev.storj.io/storj/private',
    'https://review.dev.storj.io/storj/gateway',
    'https://review.dev.storj.io/storj/gateway-mt',
    'https://review.dev.storj.io/storj/linksharing',
    'https://review.dev.storj.io/storj/storj',
    'https://review.dev.storj.io/storj/drpc'
]

def repositoryCheckStages = repositories.collectEntries {
    [ "${basename(it)}" : checkRepository(basename(it), it) ]
}

def checkRepository(name, repo) {
    return {
        stage("${name}") {
            sh "git clone --depth 2 ${repo} ${name}"
            dir(name){
                sh 'check-mod-tidy'
                sh 'check-copyright'
                sh 'check-large-files'
                sh 'check-imports ./...'
                sh 'check-peer-constraints'
                // Currently protobuf uses different structure in storj/storj repostitory,
                // making these fail.
                // sh 'storj-protobuf --protoc=$HOME/protoc/bin/protoc lint'
                // sh 'storj-protobuf --protoc=$HOME/protoc/bin/protoc check-lock'
                sh 'check-atomic-align ./...'
                sh 'check-errs ./...'
                sh 'check-monkit ./...'
                sh 'staticcheck ./...'
                // sh 'check-downgrades'

                sh 'golangci-lint run --allow-parallel-runners --config /go/ci/.golangci.yml'
            }
        }
    }
}

def basename(path) {
    lastPath = path.lastIndexOf('/');
    if (lastPath!=-1){
        path = path.substring(lastPath+1);
    }
    return path
}

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
                sh 'check-large-files'
                sh 'check-imports ./...'
                sh 'check-peer-constraints'
                sh 'storj-protobuf --protoc=$HOME/protoc/bin/protoc lint'
                sh 'storj-protobuf --protoc=$HOME/protoc/bin/protoc check-lock'
                sh 'check-atomic-align ./...'
                sh 'check-errs ./...'
                sh 'staticcheck ./...'
                sh 'golangci-lint --config /go/ci/.golangci.yml -j=2 run'
                sh 'check-downgrades'
            }
        }

        stage('Repos') {
            steps {
                script {
                    parallel(repositoryCheckStages)
                }
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
