## How the original Go SAM project was created

Init a Go SAM project

```
sam init --runtime go1.x --name modernisation-platform-instance-scheduler
cd modernisation-platform-instance-scheduler
```

Output

```
➜  sam init --runtime go1.x --name modernisation-platform-instance-scheduler
Which template source would you like to use?
        1 - AWS Quick Start Templates
        2 - Custom Template Location
Choice: 1

Choose an AWS Quick Start application template
        1 - Instance Scheduler Example
        2 - Infrastructure event management
        3 - Multi-step workflow
Template: 1

Based on your selections, the only Package type available is Zip.
We will proceed to selecting the Package type as Zip.

Based on your selections, the only dependency manager available is mod.
We will proceed copying the template using mod.

Would you like to enable X-Ray tracing on the function(s) in your application?  [y/N]: y
X-Ray will incur an additional cost. View https://aws.amazon.com/xray/pricing/ for more details

Cloning from https://github.com/aws/aws-sam-cli-app-templates (process may take a moment)

    -----------------------
    Generating application:
    -----------------------
    Name: modernisation-platform-instance-scheduler
    Runtime: go1.x
    Architectures: x86_64
    Dependency Manager: mod
    Application Template: instance-scheduler
    Output Directory: .

    Next steps can be found in the README file at ./modernisation-platform-instance-scheduler/README.md


    Commands you can use next
    =========================
    [*] Create pipeline: cd modernisation-platform-instance-scheduler && sam pipeline init --bootstrap
    [*] Validate SAM template: sam validate
    [*] Test Function in the Cloud: sam sync --stack-name {stack-name} --watch
```

See also: https://maori.geek.nz/instance-scheduler-sam-aws-golang-quickstart-dc8b4b8c49ed
