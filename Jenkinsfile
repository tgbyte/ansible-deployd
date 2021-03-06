def tag = env.BRANCH_NAME == 'master' ? 'latest' : "${env.BRANCH_NAME.replace('/', '-')}"

node {
  stage('Checkout ansible-deployd') {
    checkout scm
  }

  docker.withTool('Docker') {
    stage('Build ansible-deployd') {
      def golang = docker.image('golang:1.8')
      golang.pull()
      golang.inside('-e GIT_COMMITTER_NAME=Anonymous -e GIT_COMMITTER_EMAIL=me@privacy.net') {
        sh '/usr/local/go/bin/go get -u github.com/gorilla/mux github.com/caarlos0/env'
        sh '/usr/local/go/bin/go build -v -o out/deployd'
      }
    }

    def image
    stage('Build ansible-deployd Docker image') {
      image = docker.build("tgbyte/ansible-deployd:${tag}", "--no-cache ${workspace}")
    }

    stage('Push ansible-deployd Docker image') {
      docker.withRegistry('https://registry.hub.docker.com', 'docker-hub') {
        image.push()
      }
    }
  }
}
