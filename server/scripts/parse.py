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

def _item(tier, enchantment, category, subcategory, leftovers, split_id):
    return {
        "tier": int(tier),
        "enchantment": enchantment,
        "category": category,
        "subcategory": subcategory,
        "leftovers": leftovers,
        "split_id": split_id,
    }

def new_item(split_id):
    tier = split_id[0][1]

    if split_id[1][-2] == "@":
        category = split_id[1][:-2]
    else:
        category = split_id[1]

    if category == "2H":
        category = category + "_" + split_id[2]

    if category == "2H_TOOL":
        category = "TOOL"

    enchantment = ""
    if len(split_id[-1]) >= 3 and split_id[-1][-2] == "@":
        enchantment = split_id[-1][-1:]
    
    leftovers = split_id[2:]

    subcategory = ""
    if len(leftovers) >= 1:
        subcategory = split_id[2]
    
    if len(split_id) > 3:
        leftovers = split_id[3:]

    if len(category) > 3 and category[-2] == "@":
        category = category[:-2]
    
    if len(subcategory) > 3 and subcategory[-2] == "@":
        subcategory = subcategory[:-2]

    return _item(tier, enchantment, category, subcategory, leftovers, split_id)

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

# validation
temp_filename = 'items_temp.csv'
write_csv(temp_filename, items)
with open(temp_filename, 'r', encoding='utf-8') as temp_csv:
    reader = csv.DictReader(temp_csv)
    for row in reader:
        expected_item = new_item(eval(row['split_id']))
        actual_item = _item(
            tier=row['tier'],
            enchantment=row['enchantment'],
            category=row['category'],
            subcategory=row['subcategory'],
            leftovers=eval(row['leftovers']),
            split_id=eval(row['split_id']),
        )
        if json.dumps(actual_item, sort_keys=True) != json.dumps(expected_item, sort_keys=True):
            print("there is one difference", expected_item, actual_item)

write_csv('items.csv', items)
