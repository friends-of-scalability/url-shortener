First, you need to create a SSM parameter with Database password:

`./create-parameter.sh <name> <type> <value>`

For example:

`./create-parameter.sh /X/prod/db/password String password`


Export an environment variable with your project_code. For example:

`export PROJECT_CODE=cto-bootcamp-student-X`

After that, you can create cloudformation stacks with the following command (inside infra directory):

`sceptre --var "vpc=<vpc-id>" --var "ssmdbpassword=/X/prod/db/password" --var "subnet0=<subnet0-id>" --var="subnet1=<subnet1-id>" --var="subnet2=<subnet2-id>" --var="studentId='<studentId>'" launch-env prod`
