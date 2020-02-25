# tool-collections
This tool used to perform validations on sig and projects, as well as do some statistics.

## command: statics show
It's used to count how many stars, subscribes and prs in `openeuler` and `src-openeuler` organization.
command help:
```$xslt
show current statics of openEuler

Usage:
  validator statics show [flags]

Flags:
  -g, --giteetoken string   the gitee token
  -h, --help                help for show
  -p, --password string     the password for huawei cloud
  -r, --showpr              whether show pr count
  -s, --showstar            whether show stars count
  -w, --showsubscribe       whether show subscribe count
  -t, --threads int32       how many threads to perform (default 5)
  -u, --user string         the username for huawei cloud
``` 
Note: if it's being used as pr statistics. it will print out all of the detailed
pr in csv format.
```$xslt
Repo, CreateAt, PR Number, Auther, State, Link
openeuler/website, 2020-02-20T19:13:55+08:00, 65, blueskycs2c, open, https://gitee.com/openeuler/website/pulls/65
openeuler/website, 2020-02-14T23:43:11+08:00, 64, ZhengyuhangHans, open, https://gitee.com/openeuler/website/pulls/64
```

## command: check sig repo
it's used to check the legality of [sig file,](https://gitee.com/openeuler/community/blob/master/sig/sigs.yaml)
if the repo defined in sig yaml file not exists, the validation would fail.

command help:
```$xslt
check repo legality in sig yaml

Usage:
  validator sig checkrepo [flags]

Flags:
  -f, --filename string     the file name of sig file
  -g, --giteetoken string   the gitee token
  -h, --help                help for checkrepo
```
## command: check owner
it's used to check whether the owner defined in the owner file exists.

command help:
check owner legacy on gitee website

Usage:
  validator owner check [flags]

Flags:
  -d, --dirname string      the folder used to search the owner file
  -f, --filename string     the file name of owner file
  -g, --giteetoken string   the gitee token
  -h, --help                help for check
