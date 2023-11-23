from django.db import models


# Create your models here.
class Chart(models.Model):
    chartId = models.AutoField(primary_key=True)
    chartName = models.CharField(max_length=1000, unique=True)
    image = models.ImageField(upload_to='charts/')

    def __str__(self):
        return '"chartName":' + str(self.chartName) + ', "image":' + str(self.image)