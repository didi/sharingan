(function($){
	
	var keyNo = 'JF' + ($.fn.jquery + Math.random()).replace(/\D/g, '');
	var jfid = 0;

	function createId() {
		return "jfid-" + jfid++;
	};

	var formatters = {};
	
	var JsonFormatter = function(options) {
		this.init(options);
    };

	JsonFormatter.prototype = {
		replacements: [
			"<span class='cls'>$1</span>",
			"<span class='cls'>$1</span><span><img></span><span>",
			"$1<span class='cls'>$2</span>",
			"$1</span><span class='cls'>$2</span>",
		],
		init: function(options) {
			this.options = $.extend({}, $.fn.jsonFormat.defaults, options);
			this.setOptions();
		},
		_replace: function(i, cls) {
			return cls ? this.replacements[i].replace("cls", cls) : this.replacements[i];
		},
		setOptions: function(options) {
			options = $.extend(this.options, options);
			this.replacers = [
				{//key
					regex: (function() {
						return options.keyUseQuote ? /("\w+")(?=\:)/g : /"(\w+)"(?=\:)/g;
					})(),
					replacement: this._replace(0, "prop")
				},
				{//value-string
					regex: /("[^"]*")(?=\,|$)/gm,
					replacement: this._replace(0, "string")
				},
				{//value-number
					regex: /([\+\-]?\d+(?:\.\d+)?)(?=\,|$)/gm,
					replacement: this._replace(0, "number")
				},
				{//布尔属性
					regex: /(true|false)(?=\,|$)/gm,
					replacement: this._replace(0, "boolean")
				},
				{//[折叠开始位置
					regex: /(\[)$/gm,
					replacement: this._replace(options.collapsible ? 1 : 0, "bracket")
				},
				{//]折叠结束位置
					regex: /^(\s*)(\])/gm,
					replacement: this._replace(options.collapsible ? 3 : 2, "bracket")
				},
				{//剩余的[
					regex: /(\[)(?!<)/g,
					replacement: this._replace(0, "bracket")
				},
				{//剩余的]
					regex: /(\])(?!<)/g,
					replacement: this._replace(0, "bracket")
				},
				{//{折叠开始位置
					regex: /(\{)$/gm,
					replacement: this._replace(options.collapsible ? 1 : 0, "brace")
				},
				{//}折叠结束位置
					regex: /^(\s*)(\})/gm,
					replacement: this._replace(options.collapsible ? 3 : 2, "brace")
				},
				{//剩余的{
					regex: /(\{)(?!<)/g,
					replacement: this._replace(0, "brace")
				},
				{//剩余的}
					regex: /(\})(?!<)/g,
					replacement: this._replace(0, "brace")
				},
				{//逗号
					regex: /(\,)/g,
					replacement: this._replace(0, "comma")
				},
			];
		},
		formatToHtml: function(json) {
			var formattedJson = this.formatToJson(json);
			for (r in this.replacers) {
				//逐一替换对应文本(基于正则匹配, 如果属性中有特殊符号format可能失败)
				var replacer = this.replacers[r];
				formattedJson = formattedJson.replace(replacer.regex, replacer.replacement);
			}
			return "<pre>" + formattedJson + "</pre>";
		},
		formatToJson: function(json) {
			if (typeof json == 'string') {
				json = $.parseJSON(json);
			}
			return JSON.stringify(json, this.options.replacer, this.options.padding);
		}
	};

	$.fn.jsonFormat = function(options) {
		return this.each(function() {
			var jfId = this[keyNo];
			//检查是否已经创建了formatter
			var formatter;
			if (!jfId) {
				this[keyNo] = jfId = createId();
				formatters[jfId] = formatter = new JsonFormatter(options);
			} else {
				formatter = formatters[jfId];
				formatter.setOptions(options);
			}
			var $this = $(this);
			var json = $this.val();
			if (json) {
				//表单元素时直接去value值, data无效
				$this.val(formatter.formatToJson(json));
			} else {
				options = formatter.options;
				json = options.data;
				if (!json) {
					//还没有设置data
					return;
				}
				$this.empty().html(formatter.formatToHtml(json));
				//着色
				$.each(options.color, function(k, v) {
					$this.find("." + k).css("color", v);
				});
				$.each(options.bolds, function(i, n) {
					$this.find("." + n).css({"font-weight": "bold"});
				});
				if (options.collapsible) {//启动折叠后给图片添加样式和事件
					$this.find("img").css({"cursor": "pointer", "margin-bottom": "-2px"}).attr("src", options.expanded);;
					if (!this.collapsible) {//避免重复添加事件
						this.collapsible = true;
						$this.delegate("img", "click", function() {
							this.src = $(this).parent().next().toggle().is(':visible') ? options.expanded : options.collapsed;
						});
					}
				}	
			}
		});
	};

	$.fn.jsonFormat.defaults = {
		//缩进的空格个数
		padding: 4,
		//JSON.stringify回调
		replacer: null,
		//key是否使用引号
		keyUseQuote: true,
		//是否可折叠
		collapsible: true,
		data: null,
		//简单的颜色配置, 属性对应span的class
		//prop、string、number、boolean、comma、bracket、brace分别对应json key、字符串属性、数字属性、布尔属性、逗号、中括号、大括号
		color: {
			prop: "#92278f",
			string: "#008000",
			number: "#25aae2",
			boolean: "#f98280",
			comma: "#000000"
		},
		//加粗
		bolds: ['brace', 'bracket', 'prop', 'boolean', 'comma'],
		expanded: "expanded.gif",
		collapsed: "collapsed.gif"
	};
}(jQuery));