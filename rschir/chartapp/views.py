import hashlib
import json

from django.shortcuts import render
from django.views.decorators.csrf import csrf_exempt
from rest_framework.parsers import JSONParser
from django.http.response import JsonResponse

from chartapp.create_chart import create_chart_image
from chartapp.models import Chart
from chartapp.serializers import ChartSerializer
# Create your views here.


@csrf_exempt
def chartApi(request, id=0):
    if request.method == "GET":
        if id != 0:
            chart = Chart.objects.get(chartId=id)
            chart_serializer = ChartSerializer(chart)
            return JsonResponse(chart_serializer.data, safe=False)
        else:
            charts = Chart.objects.all()
            charts_serializer = ChartSerializer(charts, many=True)
            return JsonResponse(charts_serializer.data, safe=False)
    elif request.method == "POST":
        chart_data = JSONParser().parse(request)
        chart_hash = hashlib.sha256(json.dumps(chart_data).encode('utf-8')).hexdigest()
        # Поиск хеша в БД
        existing_chart = Chart.objects.filter(chartName=chart_hash).first()
        if existing_chart:
            # Если найден такой же хеш, то возвращаем запись из БД
            chart_serializer = ChartSerializer(existing_chart)
            return JsonResponse(chart_serializer.data, safe=False)
        else:
            # Создание картинки и сохранение ее в файл
            image_file = create_chart_image(chart_data)
            # Сохранение данных о графике и картинке в БД
            chart_serializer = ChartSerializer(data={'chartName': chart_hash, 'image': image_file})
            if chart_serializer.is_valid():
                chart_serializer.save()
                return JsonResponse("Added Successfully", safe=False)
            else:
                return JsonResponse(chart_serializer.errors, status=400)
    elif request.method == "PUT":
        chart = Chart.objects.get(chartId=id)
        chart_data = JSONParser().parse(request)
        chart_hash = hashlib.sha256(json.dumps(chart_data).encode('utf-8')).hexdigest()
        # Поиск хеша в БД
        existing_chart = Chart.objects.filter(chartName=chart_hash).first()
        if existing_chart:
            # Если найден такой же хеш, то возвращаем запись из БД
            chart_serializer = ChartSerializer(existing_chart)
            return JsonResponse(chart_serializer.data, safe=False)
        else:
            # Создание картинки и сохранение ее в файл
            image_file = create_chart_image(chart_data)
            # Сохранение данных о графике и картинке в БД
            chart_serializer = ChartSerializer(chart, data={'chartName': chart_hash, 'image': image_file})
            if chart_serializer.is_valid():
                chart_serializer.save()
                return JsonResponse("Added Successfully", safe=False)
            else:
                return JsonResponse(chart_serializer.errors, status=400)
    elif request.method == "DELETE":
        chart = Chart.objects.get(chartId=id)
        chart.delete()
        return JsonResponse("Deleted Successfully", safe=False)


