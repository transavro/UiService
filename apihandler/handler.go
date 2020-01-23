package apihandler

import (
	"context"
	"github.com/go-redis/redis"
	"github.com/micro/go-micro/util/log"
	"github.com/micro/protobuf/ptypes"
	pb "github.com/transavro/UiService/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
	"time"
)

type Server struct {
	MongoCollection *mongo.Collection
	RedisConnection  *redis.Client
}

func (h *Server) GetVendorSpecification(req *pb.UiSpecReq, stream pb.UiService_GetVendorSpecificationServer ) error {
	log.Info("Get Ui Hit ", req.Vendor, req.Brand)
	req.Vendor = strings.ToLower(req.GetVendor())
	req.Brand = strings.ToLower(req.GetBrand())
	var vendorBrandSpec pb.UiSpec
	 err := h.MongoCollection.FindOne(context.Background(), bson.D{{"vendor", req.Vendor},{"brand", req.Brand},}).Decode(&vendorBrandSpec)
	 if err != nil {
		 return err
	 }
	 err = stream.Send(&vendorBrandSpec)
	 if err != nil {
		 return err
	 }
	 return nil
}

func (h *Server) RegisterOrUpdateBrand(ctx context.Context, specification *pb.UiSpec ) (*pb.Response, error) {
	specification.Vendor = strings.ToLower(specification.GetVendor())
	specification.Brand = strings.ToLower(specification.GetBrand())
	var response pb.Response

	singleResult  := h.MongoCollection.FindOne(ctx, bson.D{{"vendor", specification.Vendor},{"brand", specification.Brand},})
	ts, _ := ptypes.TimestampProto(time.Now())
	if singleResult.Err() != nil {
		specification.CreatedAt = ts
		_, err := h.MongoCollection.InsertOne(ctx, specification)
		if err != nil {
			log.Info("inserting 1")
			return nil, err
		}

		response.ResponseMessage = "SuccessFully registered Vendor "+specification.Vendor+" and brand "+specification.Brand
		response.IsSuccessfull = true
		return &response, nil
	}
	log.Info("Trigger 3")
	var tempVendorSpec pb.UiSpec
	err := singleResult.Decode(&tempVendorSpec)
	log.Info(err)
	specification.CreatedAt = tempVendorSpec.CreatedAt
	specification.UpdatedAt = ts
	_, err = h.MongoCollection.ReplaceOne(ctx,bson.D{{"vendor", specification.Vendor},{"brand", specification.Brand},},specification)
	if err != nil {
		return nil, err
	}
	response.ResponseMessage = "SuccessFully Updated Vendor "+specification.Vendor+" and brand "+specification.Brand
	response.IsSuccessfull = true
	return &response, nil
}

func (h *Server) UnRegisterBrand(ctx context.Context, specification *pb.UiSpecReq ) (*pb.Response, error) {
	specification.Vendor = strings.ToLower(specification.GetVendor())
	specification.Brand = strings.ToLower(specification.GetBrand())

	var response pb.Response

	result := h.MongoCollection.FindOne(ctx, bson.D{{"vendor", specification.Vendor},{"brand", specification.Brand},})
	if result.Err() != nil {
		return nil, result.Err()
	}
	_, err := h.MongoCollection.DeleteOne(ctx, bson.D{{"vendor", specification.Vendor},{"brand", specification.Brand},})
	if err != nil {
		log.Info("error while deleting")
		return nil, err
	}
	response.ResponseMessage = "SuccessFully Unregistered Vendor "+specification.Vendor+" and brand "+specification.Brand
	response.IsSuccessfull = true
	return &response, nil
}
































//func (h *Server) makeMovieRows(ctx context.Context, specification *pb.VendorBrandSpecification) {
//
//	log.Info("makeMovieRows 1")
//	options := options.Find()
//	// Sort by `_id` field descending
//	options.SetSort(bson.D{{"_id", -1}})
//	log.Info("makeMovieRows 2")
//	vendorRedisKey := specification.Vendor+":"+specification.Brand
//
//	for _, rowSpec := range specification.RowSpecification {
//		rowSpec.RowName = strings.Replace(rowSpec.RowName, " ", "-", -1)
//		rowVendorRedisKey := vendorRedisKey +":"+ rowSpec.RowPosition +":"+ rowSpec.RowName
//		h.RedisConnection.SAdd(vendorRedisKey, rowVendorRedisKey)
//		cur, err := h.MovieMongoCollection.Find(ctx,bson.D{{"source", bson.D{{"$in", bson.A{rowSpec.RowSource[0], rowSpec.RowSource[1]}}}}}, options)
//		if err != nil {
//			log.Info("makeMovieRows 3")
//			log.Info(err.Error())
//		}
//
//		for cur.Next(ctx) {
//			var resultMovie movieService.MovieTile
//			err := cur.Decode(&resultMovie)
//			if err != nil {
//				log.Info("makeMovieRows 4")
//				log.Info(err)
//			}
//			resultByteArray, err := proto.Marshal(&resultMovie)
//			if err != nil {
//				log.Info("makeMovieRows 5")
//				log.Info(err)
//			}
//			saddResult := h.RedisConnection.HSet(rowVendorRedisKey, resultMovie.Tid, resultByteArray)
//			if _,err = saddResult.Result(); err != nil {
//				log.Info("makeMovieRows 6")
//				log.Info(err)
//			}
//		}
//		log.Info(cur.Close(ctx))
//	}
//}









//func (h *Server) callMovieService(specification *pb.VendorBrandSpecification) error {
//	var movieRowSpecs []*movieService.RowSpecification
//	for _, v := range specification.RowSpecification {
//		log.Info("nayan", v.RowName, v.RowShape, v.RowPosition, v.RowSource, v.ContentLanguage)
//		movieRowSpec := &movieService.RowSpecification{
//			RowSource:            v.RowSource,
//			ContentType:          v.ContentType,
//			RowName:              v.RowName,
//			RowPosition:          v.RowPosition,
//			RowShape:             v.RowShape,
//			ContentLanguage:	  v.ContentLanguage,
//		}
//		movieRowSpecs = append(movieRowSpecs, movieRowSpec)
//	}
//	log.Info(len(movieRowSpecs))
//	brandSpecification := &movieService.VendorBrandSpecification{
//		Vendor:           specification.Vendor,
//		Brand:            specification.Brand,
//		RowSpecification: movieRowSpecs,
//	}
//
//	_, err := h.MovieServiceClient.MakeMovieRows(context.TODO(), brandSpecification)
//	if err != nil {
//		return err
//	}
//	return nil
//}
