# 待辦事項AWS Lambda function
* AWS 
  * Lambda
  * API Gateway
  * DynamoDB
  
## API  
### 查詢待辦事項清單
**GET /todos?createdBy={string}&state={int}**
* Query params
  * createdBy: 待辦事項建立的人 
  * state: 0: 待處理、 1: 已完成
  
### 建立待辦事項
**POST /todo**
* Headers
  * Content-Type: application/json
* Body params
  * item: 待辦事項內容 
  * createdBy: 待辦事項建立的人

### 刪除待辦事項
**DELETE /todo?id={string}**
* Query params
  * id: 待辦事項的id 

### 更新待辦事項
**PUT todo**
* Headers
  * Content-Type: application/json
* Path params
  * id: 待辦事項的id 
* Body params
  * id: 待辦事項的id 
  * item: 待辦事項內容 
  * state: 0: 待處理、 1: 已完成
  	



