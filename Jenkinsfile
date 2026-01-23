def lastStage = ''
node('node') {
  properties([disableConcurrentBuilds()])
  try {
    currentBuild.result = "SUCCESS"

    stage('Checkout') {
      lastStage = env.STAGE_NAME

      timeout(time: 10, unit: 'MINUTES') {
        checkout scm
      }

      echo "Current build result: ${currentBuild.result}"
    }

    stage('Build and push images') {
      lastStage = env.STAGE_NAME

      env.DOCKER_BUILDKIT = '1'
      env.BUILDX_BUILDER = 'multiplatform-builder'
      sh 'docker buildx create --name $BUILDX_BUILDER --driver docker-container --bootstrap --use || docker buildx use $BUILDX_BUILDER'

      timeout(time: 2, unit: 'HOURS') {
        sh 'make build-and-push-images'
      }

      echo "Current build result: ${currentBuild.result}"
    }
  }
  catch (err) {
    echo "Caught errors! ${err}"
    echo "Setting build result to FAILURE"
    currentBuild.result = "FAILURE"

    slackSend color: 'danger', message: "@build-team ci branch ${env.BRANCH_NAME} build failed during stage ${lastStage} ${env.BUILD_URL}"

    mail from: 'builds@storj.io',
      replyTo: 'builds@storj.io',
      to: 'builds@storj.io',
      subject: "storj/ci branch ${env.BRANCH_NAME} build failed",
      body: "Project build log: ${env.BUILD_URL}"

    throw err
  }
  finally {
    stage('Cleanup') {
      sh script: '[ -n "$BUILDX_BUILDER" ] && docker buildx rm --keep-state $BUILDX_BUILDER || true', returnStatus: true
      sh script: 'make clean', returnStatus: true
      deleteDir()
    }
  }
}
