import json
from collections import Counter

projects = ['gin','echo','mux','cobra','viper','testify','logrus','redis','grpc-go','client_golang']

rule_desc = {
    'G101':'硬编码凭证','G102':'绑定所有网卡','G103':'unsafe 包','G104':'未处理错误',
    'G107':'SSRF URL','G108':'pprof 暴露','G109':'int32 溢出','G110':'zip slip',
    'G114':'不安全 http 配置','G115':'整数转换溢出','G118':'未知','G123':'未知','G124':'未知',
    'G201':'SQL fmt','G202':'SQL concat','G203':'html 未转义','G204':'subprocess',
    'G301':'目录权限','G302':'文件权限','G303':'tempfile','G304':'Path Traversal',
    'G305':'zip slip','G306':'文件写权限',
    'G401':'弱加密','G402':'TLS 跳过校验','G403':'RSA<2048','G404':'math/rand',
    'G501':'md5','G502':'des','G503':'rc4','G504':'cgi','G505':'sha1',
    'G601':'range 指针','G703':'未知','G706':'未知',
}

# 重新分类（v2，更严格）：
# 兜底层相关 = 只包括"语言/运行时本可避免但选择不阻止的"——unsafe 包、range 循环 pointer
# 开发者必守层 = 其余所有
L1 = {'G103','G601','G504'}  # 兜底层相关：语言层面开放给你的工具
# G104 未处理错误 = 工程卫生，不算严格安全，单独拎出来

def bucket(rid):
    if rid == 'G104': return 'hygiene'  # 代码卫生
    if rid in L1: return 'layer1'
    return 'layer2'

results = {}
all_rules = Counter()
bucket_counter = Counter()
severity_counter = Counter()
high_bucket_counter = Counter()

for p in projects:
    try:
        d = json.load(open(f'/tmp/gosec-{p}.json'))
    except Exception as e:
        continue
    issues = d.get('Issues', [])
    stats = d.get('Stats', {})
    per_rule = Counter()
    per_sev = Counter()
    per_bucket = Counter()
    for it in issues:
        rid = it.get('rule_id','?')
        sev = it.get('severity','?')
        per_rule[rid]+=1
        per_sev[sev]+=1
        per_bucket[bucket(rid)] += 1
        all_rules[rid]+=1
        severity_counter[sev]+=1
        bucket_counter[bucket(rid)] += 1
        if sev == 'HIGH':
            high_bucket_counter[bucket(rid)] += 1
    results[p] = {
        'issues_total': len(issues),
        'files': stats.get('files', 0),
        'lines': stats.get('lines', 0),
        'by_rule': dict(per_rule.most_common()),
        'by_severity': dict(per_sev),
        'by_bucket': dict(per_bucket),
    }

print("="*70)
print("GOSEC 扫描结果汇总（10 个主流 Go 开源项目，2026-04-30，v2 分类）")
print("="*70)
print()
print("【全部 issues 按桶分类】")
total = sum(bucket_counter.values())
for b in ['layer1','layer2','hygiene']:
    c = bucket_counter.get(b, 0)
    pct = c*100/total if total else 0
    label = {'layer1':'兜底层相关(unsafe/range指针/cgi)','layer2':'开发者必守层(注入/crypto/权限/SSRF/整数溢出/path)','hygiene':'代码卫生(未处理错误 G104)'}
    print(f"  {label[b]:<60} {c:>5} 次  {pct:.1f}%")

print()
print("【HIGH 严重度 issues 按桶（这才是真正的安全问题）】")
total_high = sum(high_bucket_counter.values())
for b in ['layer1','layer2','hygiene']:
    c = high_bucket_counter.get(b, 0)
    pct = c*100/total_high if total_high else 0
    label = {'layer1':'兜底层相关','layer2':'开发者必守层','hygiene':'代码卫生'}
    print(f"  {label[b]:<25} {c:>5} 次  {pct:.1f}%")

print()
print("【最常见规则 TOP 15】")
for rid, cnt in all_rules.most_common(15):
    b = bucket(rid)
    label = {'layer1':'兜底层','layer2':'必守层','hygiene':'卫生'}
    print(f"  {rid:<6} {rule_desc.get(rid,'?'):<25}  {cnt:>4}  [{label[b]}]")

# 存 JSON
summary = {
    'projects': results,
    'total_by_rule': dict(all_rules.most_common()),
    'total_by_bucket': dict(bucket_counter),
    'total_by_severity': dict(severity_counter),
    'high_by_bucket': dict(high_bucket_counter),
}
with open('/Users/wujiachen/WriteCraft/articles/go-security-programming/evidence/data/gosec-distribution.json','w') as f:
    json.dump(summary, f, ensure_ascii=False, indent=2)
print()
print("已保存到 evidence/data/gosec-distribution.json")
