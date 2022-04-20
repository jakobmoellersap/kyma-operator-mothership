# kyma-operator-mothership

This project hosts a golang module for the mothership operator.
This operator reacts on k8s APIs defined within the project, but it's also capable of creating/updating k8s types imported from other projects.

Steps:

1) Setup the github repository

2) Create subdirectory for the operator
    mkdir operator

3) Generate the operator project
    cd operator
    go mod init
    kubebuilder init --domain kyma-project.io
**Note: Generate both the Resource and the Controller**
    kubebuilder create api --group inventory --version v1alpha1 --kind Kyma
    make manifests

4) Push changes to github

5) Write the code that adds integration with IstioConfiguration objects
- Create of Kyma object should result in creation of IstioConfiguration
- Delete of Kyma object should result in deletion of related IstioConfiguration (if any)


