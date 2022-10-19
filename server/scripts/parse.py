import json
import csv

def write_csv(filename, data):
    with open(filename, 'w', encoding='utf8', newline='') as output_file:
        fc = csv.DictWriter(output_file, fieldnames=data[0].keys())
        fc.writeheader()
        fc.writerows(data)

def new_shopcategory(shopCategory):
    return {
        'id': shopCategory['@id'],
        'value': shopCategory['@value'],
        'shopsubcategory': shopCategory['shopsubcategory'],
        'count': len(shopCategory['shopsubcategory'])
    }

def new_item(split_id):
    tier = int(split_id[0][1])

    if split_id[1][-2] == "@":
        category = split_id[1][:-2]
    else:
        category = split_id[1]

    enchantment = ""
    if len(split_id[-1]) >= 3 and split_id[-1][-2] == "@":
        enchantment = split_id[-1][-1:]

        if len(split_id[-1]) == 3:
            split_id.pop()
        else:
            old = split_id.pop()
            split_id.append(old[:-2])
    
    leftovers = split_id[2:]

    subcategory = ""
    if len(leftovers) >= 1:
        subcategory = split_id[2]
    
    if len(split_id) > 3:
        leftovers = split_id[3:]

    return {
        "tier": tier,
        "enchantment": enchantment,
        "category": category,
        "subcategory": subcategory,
        "leftovers": leftovers,
        "original": split_id,
    }

def get_shop_category(shop_categories, split_id):
    return ""

def is_tiered(split_id):
    return len(split_id[0]) == 2 and split_id[0][0] == "T" and split_id[0][1].isnumeric()


# items by shop categories
with open('items.json', 'r', encoding='utf-8') as json_file:
    items_json_dict = json.load(json_file)

shop_categories = []
for key in items_json_dict['items'].keys():
    if key == '@xmlns:xsi' or key == '@xsi:noNamespaceSchemaLocation':
        continue

    if key == 'shopcategories':
        for shopCategory in items_json_dict['items'][key]['shopcategory']:
            shop_categories.append(new_shopcategory(shopCategory))

write_csv('shop_categories.csv', shop_categories)

# items by ID

with open('i18n.json', 'r', encoding='utf-8') as json_file:
    i18n_json_dict = json.load(json_file)

items = []
for item in i18n_json_dict:
    split_id = item['UniqueName'].split('_')
        
    if is_tiered(split_id):
        inferred_category = get_shop_category(shop_categories, split_id)
        items.append(new_item(split_id))

write_csv('items.csv', items)
