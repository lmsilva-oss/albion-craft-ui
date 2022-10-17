import json
import csv

def new_item(id, tier, category):
    return {
        "id": id,
        "tier": tier,
        "category": category,
    }

with open("items.json", "r", encoding="utf-8") as json_file:
    json_dict = json.load(json_file)

items = []

for item in json_dict:
    id = item['UniqueName']
    split_id = id.split('_')
        
    if len(split_id[0]) == 2 and split_id[0][0] == "T" and split_id[0][1].isnumeric():
        tier = int(split_id[0][1])

        if split_id[1][-2] == "@":
            category = split_id[1][:-2]
        else:
            category = split_id[1]

        items.append(new_item(id, tier, category))

with open('items.csv', 'w', encoding='utf8', newline='') as output_file:
    fc = csv.DictWriter(output_file, 
                        fieldnames=items[0].keys(),

                       )
    fc.writeheader()
    fc.writerows(items)
