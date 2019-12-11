pipeline {
  agent {
    kubernetes {
      label 'kaniko'
      yaml """
kind: Pod
metadata:
  name: kaniko
spec:
  containers:
  - name: kaniko
    image: gcr.io/kaniko-project/executor:debug
    imagePullPolicy: Always
    command:
    - /busybox/cat
    tty: true
    volumeMounts:
      - name: jenkins-docker-cfg
        mountPath: /root/.docker
      - name: kaniko-cache
        mountPath: /cache
  volumes:
  - name: kaniko-cache
    persistentVolumeClaim:
      claimName: kaniko-cache
  - name: jenkins-docker-cfg
    projected:
      sources:
      - secret:
          name: regcred
          items:
            - key: config.json
              path: config.json
"""
    }
  }
  stages {
    stage('Build with Kaniko') {
      environment {
        PATH = "/busybox:/kaniko:$PATH"
      }
      steps {
        checkout scm
        container(name: 'kaniko', shell: '/busybox/sh') {
            sh '''#!/busybox/sh
            find .
            set
            /kaniko/executor --context=dir://${WORKSPACE} --dockerfile=Dockerfile --cache=true \
              --destination=tribitio/content-filter-controller:${BUILD_TAG} \
              --destination=tribitio/content-filter-controller:latest
            '''
        }
      }
    }
  }
}
