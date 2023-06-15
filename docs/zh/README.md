# InspectMe

# 简介

InspectMe 是一个面向数据，声明式的巡检工具。它允许用户编写一个基于yaml配置去定一个巡检测试内容。

# 热身
假设你想对一组IP地址定时巡检其HTTP功能。大体上，你需要如下两个步骤

## 定义巡检IP列表
需要把IP列表按照格式存成一个JSON数组，然后存放在文件或者OSS上，让InspectMe去拉取即可。比如，将需要巡检的IP列表如下存放：

```
[
   {
      "name" : "ip1",
      "ip": "1.2.3.4",
      "port": 80
   }
   {
     "name": "ip2",
     "ip": "1.2.3.5",
     "port": 80
   }
]
```

这样，就写好了一份IP列表。

## 定义巡检task

现在我们需要编写一个yaml文件，用于定义我们想怎么巡检这个巡检列表。假设这个巡检列表在文件中

```

name: MyTest # 一个唯一的名字，用于表示您的巡检任务
comment: test local variable is cleared or not # 一段描述醒的注释

target:
  fetch:
    uri: file://test/assets/target.json
  format: json_v1 # json_v1，表示格式为JSON

# 定义触发器
trigger: trigger.Cron("@every 30s") # 定义每30s触发一次

#  定义巡检task
task:
  # 我们的task包含一个子task，即http task，hi-doctor目前支持http/oss_get/oss_put/code四种，将来会加入其他类型
  - type: http
    option:
      method: GET         # 请求方法为GET
      path: /index.html   # 请求路径
      header:             # 请求的头，如果需要，这里列出的头会加入请求的头重
        user-agent: 'curl/7.52.1'
        accept: '*/*'
      body:               #如果有任何body，可以填写字符串表示upload的body大小
      host: www.sina.com.cn #host头

    # -------------------------------------------------------------------------
    # 验证部分代码，验证区域允许用户使用表达式语法来编写希望验证的结果
    #
    # 一般而言，验证的部分由下面几个部分组成
    #  1） condition 部分，表示验证的条件表达式
    #  2） Otherwise部分，一个代码列表，当Condition的表达式执行结果为false，
    #      注意他是个yaml的List，只能包含字符串 这个部分的表达式依次执行。
    #  3） Then部分，一个代码列表，当Condition的表达式执行结果为true，
    #      注意他是个yaml的List，只能包含字符串 这个部分的表达式依次执行。
    #  4)  Lastly，同样是个表达式列表，这个部分无论如何都会执行，给用户机会
    #      做一些cleanup工作

    # 下面的代码表示，当回复状态码是202，并且回复头不包含x-done才算结束
    # -------------------------------------------------------------------------
    check:
      condition: http.resp_status == 202 and http.HeaderHas(http.resp_header, "x-done")
      otherwise:
        # 写一条日志
        - log.Error("the target %s failed", target.name)
      then:
        - log.Info("the target %s is done", target.name)
      lastly:
        - log.Info("the test for %s:%d is done", target.ip, target.port)

# 当某次触发的所有的task都执行结束后，那么finally部分的表达式被依次执行
finally:
  - log.Info("All node tests are done")
 ```

上面定义了个一个针对http请求的巡检。注意我们的巡检的ip列表包含两个IP，每次cron trigger触发的时候，那么上面的yaml 定义的task会分别对这两个IP都执行。

## 增加metrics

我们往往需要针对巡检结果进行打点，hi-doctor内置了metrics功能

```

name: MyTest # 一个唯一的名字，用于表示您的巡检任务
comment: test local variable is cleared or not # 一段描述醒的注释

target:
  fetch:
    uri: file://test/assets/target.json
  format: json_v1 # json_v1，表示格式为JSON

# 定义metrics, InspectMe会自动分配关联这些定义的metrics
# 注意，用户可以利用metrics API自己注册自己的metrics服务
metrics:
  probvider: local
  namespace: my_cool_local_task # metrices的前缀
  define:
    - name: metrics_done       # 变量名称，用于在表达式中使用该metrics
      key:  "done_counter"     # metrics的名称
      tpye: "counter"          # 类型，目前支持counter 和 gauge

    - name: metrics_fail       # 变量名称，用于在表达式中使用该metrics
      key:  "fail_counter"     # metrics的名称
      tpye: "counter"          # 类型，目前支持counter 和 gauge


# 定义触发器
trigger: trigger.Cron("@every 30s") # 定义每30s触发一次

#  定义巡检task
task:
  # 我们的task包含一个子task，即http task，hi-doctor目前支持http/oss_get/oss_put/code四种，将来会加入其他类型
  - type: http
    option:
      method: GET         # 请求方法为GET
      path: /index.html   # 请求路径
      header:             # 请求的头，如果需要，这里列出的头会加入请求的头重
        user-agent: 'curl/7.52.1'
        accept: '*/*'
      body:               #如果有任何body，可以填写字符串表示upload的body大小
      host: www.sina.com.cn #host头

    # -------------------------------------------------------------------------
    # 验证部分代码，验证区域允许用户使用表达式语法来编写希望验证的结果
    #
    # 一般而言，验证的部分由下面几个部分组成
    #  1） condition 部分，表示验证的条件表达式
    #  2） Otherwise部分，一个代码列表，当Condition的表达式执行结果为false，
    #      注意他是个yaml的List，只能包含字符串 这个部分的表达式依次执行。
    #  3） Then部分，一个代码列表，当Condition的表达式执行结果为true，
    #      注意他是个yaml的List，只能包含字符串 这个部分的表达式依次执行。
    #  4)  Lastly，同样是个表达式列表，这个部分无论如何都会执行，给用户机会
    #      做一些cleanup工作
    #
    # 下面的代码表示，当回复状态码是202，并且回复头不包含x-done才算结束
    # -------------------------------------------------------------------------
    check:
      condition: http.resp_status == 202 and http.HeaderHas(http.resp_header, "x-done")
      otherwise:
        # 写一条日志
        - log.Error("the target %s failed", target.name)
        # 记录一个metrics，所有定义的metrics指标变量都在metrics namespace下
        # 我们只解释用Emit方法即可，Emit方法的第二个参数是个map允许用户加入
        # tag, 注意 ':'是不合法的yaml语言，因此要么给整个代码加引号，要么
        # heredoc
        - >
          metrics.metrics_fail.Emit(1, {"tag1": "val1", "tag2": "val2"})
      then:
        - log.Info("the target %s is done", target.name)
        # 记录一个metrics，所有定义的metrics指标变量都在metrics namespace下
        # 我们只解释用Emit方法即可，Emit方法的第二个参数是个map允许用户加入
        # tag, 注意 ':'是不合法的yaml语言，因此要么给整个代码加引号，要么
        # heredoc
        - >
          metrics.metrics_done.Emit(1, {"tag1": "val1", "tag2": "val2"})
      lastly:
        - log.Info("the test for %s:%d is done", target.ip, target.port)

# 当某次触发的所有的task都执行结束后，那么finally部分的表达式被依次执行
finally:
  - log.Info("All node tests are done")
 ```

有时候，你需要保存下变量最后统计下结果。比如，你测试了10个节点，你希望把失败和成功的节点数目分别汇报给飞书，或者发email，这个就需要变量功能来记录测试的结果。InspectMe内置了三种不同生命周期的变量类型，分别如下

1. Storage，该变量和该Inspection Job同生命周期，永远存在，可以用于记录跨Cron的状态
2. Global，该变量每次Cron触发的时候初始化，某个CronJob结束后，销毁，用于记录跨Task状态
3. Local，该变量在每个TaskBatch开始时候出事话，TaskBatch结束后销毁。TaskBatch为定义在yaml task下面的*所有的*的task的一次执行，每次执行就是对应你target json中定义的某个巡检目标ip或者oss的请求path。

为了完成上面的问题，即记录下有多少个成功和失败，然后上报飞书。我们需要的是Global对象，因为他是跨Task存活，每次Cron触发重新初始化的。

## 添加Global状态

```

name: MyTest # 一个唯一的名字，用于表示您的巡检任务
comment: test local variable is cleared or not # 一段描述醒的注释

target:
  fetch:
    uri: file://test/assets/target.json
  format: json_v1 # json_v1，表示格式为JSON

# 定义Global状态
global:
  total_success: 0 # 初始化0， 总共成功
  total_fail: 0  # 初始化0， 总共失败


# 定义metrics, InspectMe会自动分配关联这些定义的metrics
# 注意，用户可以利用metrics API自己注册自己的metrics服务
metrics:
  probvider: local
  namespace: my_cool_local_task # metrices的前缀
  define:
    - name: metrics_done       # 变量名称，用于在表达式中使用该metrics
      key:  "done_counter"     # metrics的名称
      tpye: "counter"          # 类型，目前支持counter 和 gauge

    - name: metrics_fail       # 变量名称，用于在表达式中使用该metrics
      key:  "fail_counter"     # metrics的名称
      tpye: "counter"          # 类型，目前支持counter 和 gauge


# 定义触发器
trigger: trigger.Cron("@every 30s") # 定义每30s触发一次

#  定义巡检task
task:
  # 我们的task包含一个子task，即http task
  # hi-doctor目前支持http/oss_get/oss_put/code四种，将来会加入其他类型
  - type: http
    option:
      method: GET         # 请求方法为GET
      path: /index.html   # 请求路径
      header:             # 请求的头，如果需要，这里列出的头会加入请求的头重
        user-agent: 'curl/7.52.1'
        accept: '*/*'
      body:               #如果有任何body，可以填写字符串表示upload的body大小
      host: www.sina.com.cn #host头

# -------------------------------------------------------------------------
# 验证部分代码，验证区域允许用户使用表达式语法来编写希望验证的结果
#
# 一般而言，验证的部分由下面几个部分组成
#  1） condition 部分，表示验证的条件表达式
#  2） Otherwise部分，一个代码列表，当Condition的表达式执行结果为false，
#      注意他是个yaml的List，只能包含字符串 这个部分的表达式依次执行。
#  3） Then部分，一个代码列表，当Condition的表达式执行结果为true，
#      注意他是个yaml的List，只能包含字符串 这个部分的表达式依次执行。
#  4)  Lastly，同样是个表达式列表，这个部分无论如何都会执行，给用户机会
#      做一些cleanup工作

# 下面的代码表示，当回复状态码是202，并且回复头不包含x-done才算结束
# -------------------------------------------------------------------------
    check:
      condition: http.resp_status == 202 and http.HeaderHas(http.resp_header, "x-done")
      otherwise:
        # 写一条日志
        - log.Error("the target %s failed", target.name)
        # 记录一个metrics，所有定义的metrics指标变量都在metrics namespace下
        # 我们只解释用Emit方法即可，Emit方法的第二个参数是个map允许用户加入
        # tag, 注意 ':'是不合法的yaml语言，因此要么给整个代码加引号，要么
        # heredoc
        - >
          metrics.metrics_fail.Emit(1, {"tag1": "val1", "tag2": "val2"})

        # 打印下当前的total_fail
        - log.Info("total_fail %d", global.total_fail)

        # 设置total_fail增加
        - var.SetGlobal("total_fail", global.total_fail+1)
      then:
        - log.Info("the target %s is done", target.name)
        # 记录一个metrics，所有定义的metrics指标变量都在metrics namespace下
        # 我们只解释用Emit方法即可，Emit方法的第二个参数是个map允许用户加入
        # tag, 注意 ':'是不合法的yaml语言，因此要么给整个代码加引号，要么
        # heredoc
        - >
          metrics.metrics_done.Emit(1, {"tag1": "val1", "tag2": "val2"})

        # 打印下当前的 total_success
        - log.Info("total_success %d", global.total_success)

        # 设置total_success增加
        - var.SetGlobal("total_success", global.total_success+1)
      lastly:
        - log.Info("the test for %s:%d is done", target.ip, target.port)


# 当某次触发的所有的task都执行结束后，那么finally部分的表达式被依次执行
finally:
  - log.Info("All node tests are done")
  - log.Info("success %d, fail %d", global.total_success, global.total_fail)
 ```


# 其他Task
