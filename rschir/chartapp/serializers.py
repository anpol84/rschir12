from rest_framework import serializers
from chartapp.models import Chart

class ChartSerializer(serializers.ModelSerializer):
    image = serializers.FileField()

    class Meta:
        model = Chart
        fields = ('chartId', 'chartName', 'image')