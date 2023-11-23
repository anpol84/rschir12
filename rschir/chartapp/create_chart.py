import io
import base64
from django.core.files.base import ContentFile
from django.core.files.uploadedfile import InMemoryUploadedFile
from matplotlib import pyplot as plt


def create_chart_image(chart_data):
    # Получение данных для построения графика
    files = chart_data['chartName']['files']

    names = [file['name'] for file in files]
    descriptions = [file['description'] for file in files]

    # Построение графика
    fig, ax = plt.subplots()
    ax.bar(names, descriptions)
    ax.set_xlabel('Name')
    ax.set_ylabel('Description')
    ax.set_title('My Chart')

    # Преобразование графика в картинку и сохранение ее в виде base64-строки
    buffer = io.BytesIO()
    fig.savefig(buffer, format='png')
    buffer.seek(0)
    image_file = InMemoryUploadedFile(buffer, None, 'chart.png', 'image/png', buffer.getbuffer().nbytes, None)
    return image_file

