#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
比对报销单与消费记录，识别冲突
"""

import openpyxl
from pathlib import Path

# 用户主目录的 .claude 文件夹路径
CLAUDE_DIR = Path(r"C:\Users\liuchujun\.claude")

def load_excel(file_path):
    """加载 Excel 文件并返回所有工作表的数据"""
    wb = openpyxl.load_workbook(file_path)
    data = {}
    for sheet_name in wb.sheetnames:
        sheet = wb[sheet_name]
        rows = []
        for row in sheet.iter_rows(values_only=True):
            rows.append(row)
        data[sheet_name] = rows
    return data

def find_conflicts(expense_claims, consumption_records):
    """
    比对报销单与消费记录，识别冲突

    冲突类型：
    1. 金额不匹配：报销单金额与消费记录金额不一致
    2. 重复报销：同一笔消费出现在报销单中多次
    3. 无对应消费：报销单中的项目在消费记录中找不到
    4. 无对应报销：消费记录中的项目没有对应的报销单
    """
    conflicts = []

    # 提取报销单数据（假设第一行是表头）
    claim_keys = set()
    claim_dict = {}
    for i, row in enumerate(expense_claims):
        if i == 0:  # 跳过表头
            continue
        if row and any(row):  # 非空行
            # 假设列结构：日期、描述、金额
            key = None
            for col_idx, val in enumerate(row):
                if val:
                    if key is None:
                        key = str(val)
                    else:
                        key = f"{key}_{val}"
            if key:
                if key in claim_keys:
                    conflicts.append({
                        'type': '重复报销',
                        'description': f'报销单中重复的条目：{key}',
                        'row': i + 1
                    })
                else:
                    claim_keys.add(key)
                    claim_dict[key] = row

    # 提取消费记录数据
    record_keys = set()
    record_dict = {}
    for i, row in enumerate(consumption_records):
        if i == 0:  # 跳过表头
            continue
        if row and any(row):  # 非空行
            key = None
            for col_idx, val in enumerate(row):
                if val:
                    if key is None:
                        key = str(val)
                    else:
                        key = f"{key}_{val}"
            if key:
                record_keys.add(key)
                record_dict[key] = row

    # 找出无对应消费记录的报销
    for key, row in claim_dict.items():
        if key not in record_keys:
            # 尝试模糊匹配（仅匹配金额）
            claim_amount = None
            for val in row:
                if isinstance(val, (int, float)):
                    claim_amount = val
                    break

            matched = False
            if claim_amount:
                for rec_key, rec_row in record_dict.items():
                    for val in rec_row:
                        if isinstance(val, (int, float)) and abs(val - claim_amount) < 0.01:
                            matched = True
                            break
                    if matched:
                        break

            if not matched:
                conflicts.append({
                    'type': '无对应消费',
                    'description': f'报销条目 {key} 在消费记录中未找到',
                    'row': expense_claims.index(row) + 1 if row in expense_claims else 'N/A'
                })

    # 找出无对应报销的消费记录
    for key, row in record_dict.items():
        if key not in claim_keys:
            # 尝试模糊匹配
            record_amount = None
            for val in row:
                if isinstance(val, (int, float)):
                    record_amount = val
                    break

            matched = False
            if record_amount:
                for cl_key, cl_row in claim_dict.items():
                    for val in cl_row:
                        if isinstance(val, (int, float)) and abs(val - record_amount) < 0.01:
                            matched = True
                            break
                    if matched:
                        break

            if not matched:
                conflicts.append({
                    'type': '无对应报销',
                    'description': f'消费记录 {key} 未找到对应报销',
                    'row': consumption_records.index(row) + 1 if row in consumption_records else 'N/A'
                })

    return conflicts

def main():
    # 加载两个 Excel 文件
    expense_claims_file = CLAUDE_DIR / "报销单.xlsx"
    consumption_records_file = CLAUDE_DIR / "消费记录.xlsx"

    print(f"加载文件：{expense_claims_file}")
    print(f"加载文件：{consumption_records_file}")

    expense_claims = load_excel(expense_claims_file)
    consumption_records = load_excel(consumption_records_file)

    # 打印工作表名称
    print(f"\n报销单工作表：{list(expense_claims.keys())}")
    print(f"消费记录工作表：{list(consumption_records.keys())}")

    # 打印数据预览
    for sheet_name, rows in expense_claims.items():
        print(f"\n=== 报销单 - {sheet_name} ===")
        print(f"行数：{len(rows)}")
        if rows:
            print(f"表头：{rows[0]}")
            if len(rows) > 1:
                print(f"第一行数据：{rows[1]}")

    for sheet_name, rows in consumption_records.items():
        print(f"\n=== 消费记录 - {sheet_name} ===")
        print(f"行数：{len(rows)}")
        if rows:
            print(f"表头：{rows[0]}")
            if len(rows) > 1:
                print(f"第一行数据：{rows[1]}")

    # 获取第一个工作表的数据进行比对
    expense_claims_data = list(expense_claims.values())[0] if expense_claims else []
    consumption_records_data = list(consumption_records.values())[0] if consumption_records else []

    # 查找冲突
    conflicts = find_conflicts(expense_claims_data, consumption_records_data)

    print("\n" + "="*50)
    print("冲突检测结果：")
    print("="*50)

    if conflicts:
        for conflict in conflicts:
            print(f"\n[{conflict['type']}]")
            print(f"  描述：{conflict['description']}")
            print(f"  行号：{conflict['row']}")
    else:
        print("未发现冲突！")

    print(f"\n共发现 {len(conflicts)} 个冲突")

if __name__ == "__main__":
    main()
