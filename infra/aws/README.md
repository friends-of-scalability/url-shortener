Export an environment variable with your project_code. For example:

`export PROJECT_CODE=cto-bootcamp-student-X`

After that, you can create cloudformation stacks with the following command (inside infra directory):

`sceptre --var "VpcName=<student-name>" launch-env bootstrap`
