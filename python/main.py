import requests
from bs4 import BeautifulSoup
from fake_useragent import UserAgent
from selenium import webdriver
from webdriver_manager.chrome import ChromeDriverManager
import os
import pandas as pd
import csv
import time
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

user_agent = UserAgent().chrome

user_agent = 'Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/28.0.1468.0 Safari/537.36'
url = "https://phimgihd.net/phim-le/"
headers = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36 QIHU 360SE'
}

def get_detail(src, detail_url):
    if src == 'phimgihd':
        response = requests.get(detail_url, headers=headers)
        soup = BeautifulSoup(response.content, 'lxml')
        name_element = soup.find('h1', {'class': 'entry-title'})
        name_content = name_element.getText() if name_element else None

        org_element = soup.find('img', {'class': 'movie-thumb'})
        org_title = org_element.getText() if org_element else None
        
        img_element = soup.find('p', {'class': 'org_title'})
        img = org_element.attrs['src'] if img_element else None

        last_eps_element = soup.find('span', {'class': 'last-eps box-shadow'})
        last_eps = last_eps_element.getText() if last_eps_element else None
       
        # year = soup.find('span', {'class': 'title-year'}).getText()
        year_element = soup.find('span', {'class': 'title-year'})
        year = year_element.getText() if year_element else None 
       
        category_element = soup.find('p', {'class': 'category'})
        category = category_element.getText() if category_element else None

        actors_element = soup.find('p', {'class': 'actors'})
        actors = actors_element.getText() if actors_element else None

        url_watch =  soup.find(
                'a', {'class': 'btn btn-sm btn-danger watch-movie visible-xs-blockx'}
            ).attrs['href']
        # Selenium setup
        os.environ['MOZ_HEADLESS'] = '1'
        opts = webdriver.ChromeOptions()
        opts.add_argument("--headless")
        opts.headless = True
        driver = webdriver.Chrome(opts)
        driver.maximize_window()
        # wait = WebDriverWait(driver, 30)
        driver.get(url_watch)
        # driver.execute_script("var scrollingElement = (document.scrollingElement || document.body);scrollingElement.scrollTop = scrollingElement.scrollHeight;")
        # wait.until(EC.frame_to_be_available_and_switch_to_it((By.CSS_SELECTOR, "iframe[name='video_player']")))
        # wait.until(EC.frame_to_be_available_and_switch_to_it((By.CSS_SELECTOR, "iframe[class='embed-responsive-item']")))
        # video_url = wait.until(EC.visibility_of_element_located((By.XPATH, "//div[@class='jw-media jw-reset']//*"))).get_attribute('src')
        soup_watch = BeautifulSoup(driver.page_source, "html.parser")
        ember_url_full = ''
        note = ''
        # driver.switch_to.default_content()
        server = soup_watch.find('span', {'class': 'get-eps no-active box-shadow active'})
        if server is None:
            servers =  soup_watch.find_all('span', {'class': 'get-eps no-active box-shadow'})
            if servers.__len__() == 0:
                server = soup_watch.find('span', {'class': 'get-eps no-active box-shadow'})
            server = servers[0]    
            
        serverstr = server.getText()
        # driver.switch_to.frame(driver.find_element(By.CLASS_NAME,"embed-responsive-item"))
        try:
            ember_url_full = soup_watch.find('iframe', {'class': 'embed-responsive-item'}).attrs['src']
        except Exception as error:
            ember_url_full = None
            note = error
            
        print(ember_url_full)
        driver.quit()
        response.close()
        return {
            "name_content": name_content,
            "org_title": org_title,
            "last_eps": last_eps,
            "year": year,
            "category": category,
            "actors": actors,
            "emberUrlFull": ember_url_full,
            "serverstr": serverstr,
            "pathcode": url_watch.split('/')[-2],
            "note": note,
            "img": img,
            "from": detail_url,
            "src": src
        }
    
    return None


response = requests.get(url, headers=headers)
movies_lst = []
soup = BeautifulSoup(response.content, 'lxml')

table = soup.find('div', {'class': 'halim_box'})
listraw = []
i = 0
if table:
    movies = table.find_all('div', 'halim-item')
    for anchor in movies:
        i = i + 1
        movie_url = anchor.find('a', {'class': 'halim-thumb'}).attrs['href']
        print(str(i) + " - processing: " + movie_url + "...")
        movie_details = get_detail("phimgihd", movie_url)
        if movie_details is None:
            continue
        listraw.append(movie_details)
        time.sleep(1)
else:
    print("Table element not found on the page.")
response.close()

filename = "output.csv"

# Extracting the keys from the first dictionary in the list
fieldnames = listraw[0].keys()

with open(filename, "w",encoding="utf-8") as csvfile:
    writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
    
    writer.writeheader()  # Write the header row
    for row in listraw:
        writer.writerow(row) 