package static

const CustomJS = `
$(function() {
	console.log("Retrieving current configuration..");
	$.get("/configuration", function(data, textStatus) {
		console.log("received:");
		console.log(data);

		var config = JSON.parse(data);
		var currentTimeBetween = config.SecondsBetweenCaptures / 60;
		var currentInitialOffset = config.OffsetWithinHour == -1 ? -1 : config.OffsetWithinHour / 60;
		var resolution = config.ResolutionSetting;
		var rotation = config.RotateBy;

		$("#frequency").val(currentTimeBetween);
		$("#offset").val(currentInitialOffset);
		$("#rotation").val(rotation);
		$("#resolution").val(resolution);
	});

	$("#saveConfigBtn").click(function() {
		console.log("Updating current configuration..");

		var timeBetweenRaw = parseInt($("#frequency").val());
		var initialOffsetRaw = parseInt($("#offset").val());
		var rotationRaw = parseInt($("#rotation").val());
		var resolutionRaw = parseInt($("#resolution").val());

		var timeBetween = 60 * timeBetweenRaw;
		var initialOffset = initialOffsetRaw == -1 ? -1 : 60 * initialOffsetRaw;

		var updatedConf = {
			timeBetween:   timeBetween,
			initialOffset: initialOffset,
			resolution:    resolutionRaw,
			rotateBy:      rotationRaw
		};
		console.log("Updating config to");
		console.log(updatedConf);

		$.ajax({
		  type: "POST",
		  url: "/configuration",
		  data: JSON.stringify(updatedConf),
		  success: function(data, textStatus) {
			console.log("received:");
		 	console.log(data);
		 	$("#changesSaved").show();
		  }
		});
	});
});
`
