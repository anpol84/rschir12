from chartapp import views
from django.urls import re_path

urlpatterns = [
    re_path(r'^charts$', views.chartApi),
    re_path(r'^charts/([0-9]+)$', views.chartApi),
    re_path(r'^polyakov$', views.echoApi)
]